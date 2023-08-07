using Google.Protobuf;
using Grpc.Core;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class SubredditTransactionManager
{
    private readonly ITransactionManager _txManager;
    private readonly ITransactionManagerConfig _config;
    private readonly GrpcClientFactory _grpcClientFactory;

    public SubredditTransactionManager(ITransactionManager txManager, ITransactionManagerConfig config,
     GrpcClientFactory grpcClientFactory)
    {
        _txManager = txManager;
        _config = config;
        _grpcClientFactory = grpcClientFactory;
    }



    public async Task<Subreddit?> CreateSubredditAsync(Subreddit subreddit)
    {
        int shardNumber = GetSubreddditShardNumber(subreddit, _config.NumberOfShards);
        var clients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            clients.Add(_grpcClientFactory.CreateClient<SubredditGRPC.SubredditGRPCClient>(
                $"{nameof(SubredditGRPC.SubredditGRPCClient)}/S{shardNumber}_R{i}"
            ));
        }

        var subredditInfo = new SubredditInfo
        {
            Subreddit = subreddit,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            }
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var SubredditClient = (SubredditGRPC.SubredditGRPCClient)client;
            var inputSubredditInfo = (SubredditInfo)inputMessage;

            return (IMessage)await SubredditClient.CreateSubredditAsync(inputSubredditInfo, cancellationToken: cancellationToken);
        }

        var txInfo = new TransactionInfo(Guid.NewGuid(), shardNumber, clients, execFunc, subredditInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo});

        return result.Length == 1 ? (Subreddit) result[0] : null;
    }


    public async Task<Subreddit?> GetSubredditAsync(string handle)
    {
        var subreddit = new Subreddit {Handle = handle};

        int shardNumber = GetSubreddditShardNumber(subreddit, _config.NumberOfShards);
        var clients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            clients.Add(_grpcClientFactory.CreateClient<SubredditGRPC.SubredditGRPCClient>(
                $"{nameof(SubredditGRPC.SubredditGRPCClient)}/S{shardNumber}_R{i}"
            ));
        }

        var subredditInfo = new SubredditInfo
        {
            Subreddit = subreddit,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            }
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var SubredditClient = (SubredditGRPC.SubredditGRPCClient)client;
            var inputSubredditInfo = (SubredditInfo)inputMessage;

            return (IMessage)await SubredditClient.GetSubredditAsync(inputSubredditInfo, cancellationToken: cancellationToken);
        }

        var txInfo = new TransactionInfo(Guid.NewGuid(), shardNumber, clients, execFunc, subredditInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo});

        return result.Length == 1 ? (Subreddit) result[0] : null;
    }

    public async Task<IEnumerable<string>> GetSubredditsHandlesAsync()
    {
        var txs = new List<TransactionInfo>();

        for (int i = 0; i < _config.NumberOfShards; i++)
        {

            var clients = new List<ClientBase>();
            for (int j = 0; j < _config.NumberOfReplicas; j++)
            {
                clients.Add(_grpcClientFactory.CreateClient<SubredditGRPC.SubredditGRPCClient>(
                    $"{nameof(SubredditGRPC.SubredditGRPCClient)}/S{i}_R{j}"
                ));
            }

            var messageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            };
            

            static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
            {
                var SubredditClient = (SubredditGRPC.SubredditGRPCClient)client;
                var inputMessageInfo = (MessageInfo)inputMessage;

                return (IMessage)await SubredditClient.GetSubredditsHandlesAsync(inputMessageInfo, cancellationToken: cancellationToken);
            }

            var txInfo = new TransactionInfo(Guid.NewGuid(), i, clients, execFunc, messageInfo);
            txs.Add(txInfo);
        }

        // no need to twopc. every tx is submitted on its own.
        var subredditsHandlesTasks = txs.Select(tx => _txManager.SubmitTransactionsAsync(new TransactionInfo[] {tx}));
        
        var txResults = await Task.WhenAll(subredditsHandlesTasks);

        if (txResults.Length != txs.Count)
            return Enumerable.Empty<string>();

        var result = new List<string>();

        foreach (var txResult in txResults)
        {
            var subList = (SubredditList)txResult[0];

            foreach (var sub in subList.Subreddits)
            {
                result.Add(sub.Handle);
            }
        }

        return result;
    }

    internal static int GetSubreddditShardNumber(Subreddit subreddit, int nShards)
    {
        // ! string.GetHashCode() is randomized in .NET core
        return subreddit.Handle.GetDeterministicHashCode().Mod(nShards);
    }
}
