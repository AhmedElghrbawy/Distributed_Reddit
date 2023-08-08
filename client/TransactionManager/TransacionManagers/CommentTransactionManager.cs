using Google.Protobuf;
using Grpc.Core;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class CommentTransactionManager
{
    private readonly ITransactionManager _txManager;
    private readonly ITransactionManagerConfig _config;
    private readonly GrpcClientFactory _grpcClientFactory;

    public CommentTransactionManager(ITransactionManager txManager, ITransactionManagerConfig config,
     GrpcClientFactory grpcClientFactory)
    {
        _txManager = txManager;
        _config = config;
        _grpcClientFactory = grpcClientFactory;
    }


     public async Task<Comment?> CreateCommentAsync(Comment comment, string subredditHandle)
     {
        var txId = Guid.NewGuid();

        int subredditShard = SubredditTransactionManager.GetSubreddditShardNumber(new Subreddit{Handle = subredditHandle}, _config.NumberOfShards);
        int userShard = UserTransactionManager.GetUserShardNumber(new User{Handle = comment.OwnerHandle}, _config.NumberOfShards);


        var subredditShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            subredditShardClients.Add(_grpcClientFactory.CreateClient<CommentGRPC.CommentGRPCClient>(
                $"{nameof(CommentGRPC.CommentGRPCClient)}/S{subredditShard}_R{i}"
            ));
        }

        var userShardClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            userShardClients.Add(_grpcClientFactory.CreateClient<CommentGRPC.CommentGRPCClient>(
                $"{nameof(CommentGRPC.CommentGRPCClient)}/S{userShard}_R{i}"
            ));
        }


        var commentInfo = new CommentInfo
        {
            Comment = comment,
            SubredditShard = subredditShard,
            UserShard = userShard,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            },
            TwopcInfo = new TwoPhaseCommitInfo
            {
                TransactionId = txId.ToString(),
            },
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var commentClient = (CommentGRPC.CommentGRPCClient)client;
            var inputCommentInfo = (CommentInfo)inputMessage;

            return (IMessage)await commentClient.AddCommentAsync(inputCommentInfo, cancellationToken: cancellationToken);
        }

        var subredditCommentTxInfo = new TransactionInfo(txId, subredditShard, subredditShardClients, execFunc, commentInfo);
        var userCommentTxInfo = new TransactionInfo(txId, userShard, userShardClients, execFunc, commentInfo);


        var txInfos = new List<TransactionInfo>(){subredditCommentTxInfo};

        if (subredditShard != userShard)
        {
            txInfos.Add(userCommentTxInfo);
        }

        var result = await _txManager.SubmitTransactionsAsync(txInfos.ToArray());

        return result.Length == txInfos.Count ? (Comment) result[0] : null;
     }

}
