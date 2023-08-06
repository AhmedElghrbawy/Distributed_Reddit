using RDB.TransactionManager;
using rdb_grpc;

namespace DI.Services;

public class PostService
{
    private readonly PostTransactionManager _txManager;

    public PostService(PostTransactionManager txManager)
    {
        _txManager = txManager;
    }

    public async Task<Post> CreatePostAsync(Post post)
    {
        return await _txManager.CreatePostAsync(post);
    }
}
