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
    private static readonly TimeSpan _transactionTimeoutPeriod = TimeSpan.FromMilliseconds(2000); 
    private static readonly TimeSpan _callTimeoutPeriod = TimeSpan.FromMilliseconds(600); 

    public TransactionManager(ITransactionManagerConfig config)
    {
        _config = config;
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

        throw new NotImplementedException(); 
    }

    /// <summary>
    ///  Loops through every replica/client in the specified transaction shard trying to submit the transaction to it. 
    /// </summary>
    /// <param name="txInfo">The transaction info</param>
    /// <param name="TransactionCt">The CancellationToken used to cancel the submition of the transaction across the shard's replica</param>
    /// <returns></returns>
    /// <exception cref="OperationCanceledException"></exception>
    private async Task<IMessage> RunTransactionAsync(TransactionInfo txInfo, CancellationToken TransactionCt)
    {
        for (int i = 0; i < txInfo.Clients.Count; i = (i + 1) % _config.NumberOfReplicas)
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
                return await txInfo.ExecutionFunction(txInfo.InputMessage, client, linkedCts.Token);
            }
            catch (RpcException ex) 
                when (ex.StatusCode == StatusCode.Unavailable
                    || ex.StatusCode == StatusCode.DeadlineExceeded
                    || ex.StatusCode == StatusCode.Cancelled)
            {
                // retry with different client
                // the server should handle duplicates
                System.Console.WriteLine(ex.Message);
                continue;
            }

        }    

        throw new UnreachableException(); 
    }

}
