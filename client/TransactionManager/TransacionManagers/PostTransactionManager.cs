using Google.Protobuf;
using Grpc.Core;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class PostTransactionManager
{
    private readonly ITransactionManager _txManager;
    private readonly ITransactionManagerConfig _config;
    private readonly GrpcClientFactory _grpcClientFactory;

    public PostTransactionManager(ITransactionManager txManager, ITransactionManagerConfig config,
     GrpcClientFactory grpcClientFactory)
    {
        _txManager = txManager;
        _config = config;
        _grpcClientFactory = grpcClientFactory;
    }

    public async Task<Post?> CreatePostAsync(Post post)
    {
        var txId = Guid.NewGuid();

        int subredditShard = SubredditTransactionManager.GetSubreddditShardNumber(new Subreddit{Handle = post.SubredditHandle}, _config.NumberOfShards);
        int userShard = UserTransactionManager.GetUserShardNumber(new User{Handle = post.OwnerHandle}, _config.NumberOfShards);

        var subredditPostShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            subredditPostShardClients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                $"{nameof(PostGRPC.PostGRPCClient)}/S{subredditShard}_R{i}"
            ));
        }

        var userPostShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            userPostShardClients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                $"{nameof(PostGRPC.PostGRPCClient)}/S{userShard}_R{i}"
            ));
        }

        var postInfo = new PostInfo
        {
            Post = post,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            },
            SubredditShard = subredditShard,
            UserShard = userShard, 
            TwopcInfo = new TwoPhaseCommitInfo {TransactionId = txId.ToString(), TwopcEnabled = true }
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var postClient = (PostGRPC.PostGRPCClient)client;
            var inputPostInfo = (PostInfo)inputMessage;

            return (IMessage)await postClient.CreatePostAsync(inputPostInfo, cancellationToken: cancellationToken);
        }

        var subredditPostTxInfo = new TransactionInfo(txId, subredditShard, subredditPostShardClients, execFunc, postInfo);
        var userPostTxInfo = new TransactionInfo(txId, userShard, userPostShardClients, execFunc, postInfo);


        var txInfos = new List<TransactionInfo>(){subredditPostTxInfo};

        // no need to 2pc if the subredditPost and userPost belongs to the same shard
        if (subredditShard != userShard)
        {
            txInfos.Add(userPostTxInfo);
        }

        var result = await _txManager.SubmitTransactionsAsync(txInfos.ToArray());

        return result.Length == txInfos.Count ? (Post) result[0] : null;
    }


    public async Task<Post?> GetPostAsync(Guid id, string subredditHandle, string ownerHandle)
    {
        int subredditShard = SubredditTransactionManager.GetSubreddditShardNumber(new Subreddit{Handle = subredditHandle}, _config.NumberOfShards);
        
        var subredditPostShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            subredditPostShardClients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                $"{nameof(PostGRPC.PostGRPCClient)}/S{subredditShard}_R{i}"
            ));
        }

        var postInfo = new PostInfo
        {
            Post = new Post {Id = id.ToString()},
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            },
            SubredditShard = subredditShard,
            UserShard = 0, 
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var postClient = (PostGRPC.PostGRPCClient)client;
            var inputPostInfo = (PostInfo)inputMessage;

            return (IMessage)await postClient.GetPostAsync(inputPostInfo, cancellationToken: cancellationToken);
        }

        var postTxInfo = new TransactionInfo(Guid.NewGuid(), subredditShard, subredditPostShardClients, execFunc, postInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {postTxInfo});

        return result.Length == 1 ? (Post) result[0] : null;
    }


    public async Task<Post?> UpdatePostAsync(Post post, IEnumerable<PostUpdatedColumn> updatedColumns)
    {
        var txId = Guid.NewGuid();

        int subredditShard = SubredditTransactionManager.GetSubreddditShardNumber(new Subreddit{Handle = post.SubredditHandle}, _config.NumberOfShards);
        int userShard = UserTransactionManager.GetUserShardNumber(new User{Handle = post.OwnerHandle}, _config.NumberOfShards);

        var subredditPostShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            subredditPostShardClients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                $"{nameof(PostGRPC.PostGRPCClient)}/S{subredditShard}_R{i}"
            ));
        }

        var userPostShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            userPostShardClients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                $"{nameof(PostGRPC.PostGRPCClient)}/S{userShard}_R{i}"
            ));
        }

        var postInfo = new PostInfo
        {
            Post = post,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            },
            SubredditShard = subredditShard,
            UserShard = userShard, 
            TwopcInfo = new TwoPhaseCommitInfo {TransactionId = txId.ToString(), TwopcEnabled = true }
        };

        postInfo.UpdatedColumns.AddRange(updatedColumns);

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var postClient = (PostGRPC.PostGRPCClient)client;
            var inputPostInfo = (PostInfo)inputMessage;

            return (IMessage)await postClient.UpdatePostAsync(inputPostInfo, cancellationToken: cancellationToken);
        }

        var subredditPostTxInfo = new TransactionInfo(txId, subredditShard, subredditPostShardClients, execFunc, postInfo);
        var userPostTxInfo = new TransactionInfo(txId, userShard, userPostShardClients, execFunc, postInfo);

        // we should use 2pc when we update, but since we only support up/down vote for now, there is no need to 2pc. we can let number of votes be inconsistent

        var txInfos = new List<TransactionInfo>(){subredditPostTxInfo, userPostTxInfo};
        var txTasks = txInfos.Select(txInfo => _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo}));

        var resultTask = await Task.WhenAny(txTasks);
        var result = await resultTask;

        return result.Length == 1 ? (Post) result[0] : null;
    }

    public async Task<IEnumerable<Post>> GetPostsAsync()
    {
        var txs = new List<TransactionInfo>();

        for (int i = 0; i < _config.NumberOfShards; i++)
        {
            var clients = new List<ClientBase>();
            for (int j = 0; j < _config.NumberOfReplicas; j++)
            {
                clients.Add(_grpcClientFactory.CreateClient<PostGRPC.PostGRPCClient>(
                    $"{nameof(PostGRPC.PostGRPCClient)}/S{i}_R{j}"
                ));
            }

            var messageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            };

            static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
            {
                var postClient = (PostGRPC.PostGRPCClient)client;
                var inputPostInfo = (MessageInfo)inputMessage;

                return (IMessage)await postClient.GetPostsAsync(inputPostInfo, cancellationToken: cancellationToken);
            }

            var txInfo = new TransactionInfo(Guid.NewGuid(), i, clients, execFunc, messageInfo);
            txs.Add(txInfo);
        }


        var postsTasks = txs.Select(tx => _txManager.SubmitTransactionsAsync(new TransactionInfo[] {tx}));
        
        var txResults = await Task.WhenAll(postsTasks);

        if (txResults.Length != txs.Count)
            return Enumerable.Empty<Post>();

        var result = new List<Post>();

        foreach (var txResult in txResults)
        {
            var postList = (PostList)txResult[0];

            foreach (var post in postList.Posts)
            {
                result.Add(post);
            }
        }

        return result;
    }

    internal static int GetPostShardNumber(Post post, int nShards)
    {
        // ! string.GetHashCode() is randomized in .NET core
        return post.Id.GetDeterministicHashCode().Mod(nShards);
    }
}
