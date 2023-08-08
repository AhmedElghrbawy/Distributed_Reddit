using RDB.TransactionManager;
using rdb_grpc;

namespace DI.Services;

public class CommentService
{
    private readonly CommentTransactionManager _txManager;

    public CommentService(CommentTransactionManager txManager)
    {
        _txManager = txManager;
    }

    public async Task<Comment?> CreateCommentAsync(Comment comment, string subredditHandle)
    {
        return await _txManager.CreateCommentAsync(comment, subredditHandle);
    }

}
