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
        int userShard = 0; // TODO

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

    internal static int GetPostShardNumber(Post post, int nShards)
    {
        // ! string.GetHashCode() is randomized in .NET core
        return post.Id.GetDeterministicHashCode().Mod(nShards);
    }
}
