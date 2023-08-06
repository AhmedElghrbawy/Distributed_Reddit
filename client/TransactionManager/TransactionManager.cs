using System.Collections.Concurrent;
using System.Diagnostics;
using System.Text;
using Google.Protobuf;
using Google.Protobuf.WellKnownTypes;
using Grpc.Core;
using Grpc.Net.Client;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class TransactionManager : ITransactionManager
{
    private readonly ITransactionManagerConfig _config;
    private readonly GrpcClientFactory _grpcClientFactory;

    private static readonly TimeSpan _transactionTimeoutPeriod = TimeSpan.FromMilliseconds(2900); 
    private static readonly TimeSpan _callTimeoutPeriod = TimeSpan.FromMilliseconds(1100);

    private readonly ConcurrentDictionary<int ,int> _shardLeader; 

    public TransactionManager(ITransactionManagerConfig config, GrpcClientFactory grpcClientFactory)
    {
        _config = config;
        _grpcClientFactory = grpcClientFactory;
        _shardLeader = new ConcurrentDictionary<int ,int>();

        for (int i = 0; i < _config.NumberOfShards; i++)
        {
            _shardLeader[i] = 0;
        }
    }
    public async Task<List<IMessage>> SubmitTransactionsAsync(List<TransactionInfo> txs)
    {       
        using var transactionCts = new CancellationTokenSource();

        var results = new List<IMessage>();
        if (txs.Count == 1)
        {
            transactionCts.CancelAfter(_transactionTimeoutPeriod);
            try
            {
                results.Add(await RunTransactionAsync(txs[0], transactionCts.Token));
                return results;
            }
            catch(OperationCanceledException ex)
            {
                System.Console.WriteLine(ex.Message);
                return results;
            }
        }      

        // two phase commit
        System.Console.WriteLine("Starting 2pc");
        // preparation phase
        // the server side is aware of transaction that should be prepared
        var preparedTransactionsTasks = txs.Select(tx => RunTransactionAsync(tx, transactionCts.Token));
        IMessage[] preparedTxsResults = Array.Empty<IMessage>();

        transactionCts.CancelAfter(_transactionTimeoutPeriod);
        try
        {
            preparedTxsResults = await Task.WhenAll(preparedTransactionsTasks);
        }
        catch(OperationCanceledException ex)
        {
            // should start a task in the background to abort these transactions
            System.Console.WriteLine("preparing tx error: " + ex.Message);
            
            foreach (var tx in txs)
            {
                // no need to await here
                // we can report failure to the client and this will run in the background
                RollbackPreparedTransactionAsync(tx.TransactionId, tx.ShardNumber);
            }
        
            return preparedTxsResults.ToList();
        }
        System.Console.WriteLine("prepareation phase done");

        // commitment phase

        var commitedTransactionsTasks = txs.Select(tx => CommitPreparedTransactionAsync(tx.TransactionId, tx.ShardNumber));

        await Task.WhenAll(commitedTransactionsTasks);
        
        
        return preparedTxsResults.ToList();
    }

    /// <summary>
    ///  Loops through every replica/client in the specified transaction shard trying to submit the transaction to it. <para />
    ///  Passing a CancellationToken simulates a best-effort netwrok.<para />
    ///  Passing CancellationToken.None simulates a persistent sender (at-least-once or exactly-once depending on duplicate detection on the server side).<para />
    ///  In case of CancellationToken.None, the method is guaranteed to return after sometime (assuming a majority of servers is still up and running).
    /// </summary>
    /// <param name="txInfo">The transaction info</param>
    /// <param name="TransactionCt">The CancellationToken used to cancel the submition of the transaction across the shard's replica</param>
    /// <returns></returns>
    /// <exception cref="OperationCanceledException"></exception>
    private async Task<IMessage> RunTransactionAsync(TransactionInfo txInfo, CancellationToken TransactionCt)
    {
        // optimization: cache each shard leader number to avoid unnecessary calls.
        for (int i = _shardLeader[txInfo.ShardNumber]; i < txInfo.Clients.Count; i = (i + 1) % _config.NumberOfReplicas)
        {
            if (TransactionCt.IsCancellationRequested)
            {
                System.Console.WriteLine("Transaction cancelled");
                TransactionCt.ThrowIfCancellationRequested();
            }

            // used for cancelling this single call for the current replica
            using var singleCallCts = new CancellationTokenSource();
            singleCallCts.CancelAfter(_callTimeoutPeriod);

            // a call should be cancelled if its timout duration passes, or if the whole transction is canceled
            using var linkedCts = CancellationTokenSource.CreateLinkedTokenSource(singleCallCts.Token, TransactionCt);


            var client = txInfo.Clients[i];
            try 
            {
                var result = await txInfo.ExecutionFunction(txInfo.InputMessage, client, linkedCts.Token);

                // cache the leader number for this shard
                _shardLeader[txInfo.ShardNumber] = i;
                return result;
            }
            catch (RpcException ex) 
                when (ex.StatusCode == StatusCode.Unavailable
                    || ex.StatusCode == StatusCode.DeadlineExceeded
                    || ex.StatusCode == StatusCode.Cancelled)
            {
                // retry with different client
                // the server should handle duplicates
                System.Console.WriteLine($"Single call for S{txInfo.ShardNumber}_R{i} error: " + ex.Message);
                continue;
            }

        }    

        throw new UnreachableException(); 
    }

    private async Task CommitPreparedTransactionAsync(Guid transactionId, int shardNumber)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var twopcClient = (TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient)client;
            var inputTwopcInfo = (TwoPhaseCommitInfo)inputMessage;

            return (IMessage)await twopcClient.CommitAsync(inputTwopcInfo, cancellationToken: cancellationToken);
        }

        await RunCommitmentPhaseAsync(transactionId, shardNumber, execFunc);
    }

    private async Task RollbackPreparedTransactionAsync(Guid transactionId, int shardNumber)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var twopcClient = (TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient)client;
            var inputTwopcInfo = (TwoPhaseCommitInfo)inputMessage;

            return (IMessage)await twopcClient.RollbackAsync(inputTwopcInfo, cancellationToken: cancellationToken);
        }

        await RunCommitmentPhaseAsync(transactionId, shardNumber, execFunc);
    }


    private async Task RunCommitmentPhaseAsync(Guid transactionId, int shardNumber, Func<IMessage, ClientBase, CancellationToken, Task<IMessage>> execFunc)
    {
        var clients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            clients.Add(_grpcClientFactory.CreateClient<TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient>(
                $"{nameof(TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient)}/S{shardNumber}_R{i}"
            ));
        }

        var twopcInfo = new TwoPhaseCommitInfo
        {
            TransactionId = transactionId.ToString(),
            TwopcEnabled = true
        };


        var txInfo = new TransactionInfo(transactionId, shardNumber, clients, execFunc, twopcInfo);
        await RunTransactionAsync(txInfo, CancellationToken.None);
    }
}
