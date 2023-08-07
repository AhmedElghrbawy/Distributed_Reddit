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

    public async Task<Post?> CreatePostAsync(Post post)
    {
        return await _txManager.CreatePostAsync(post);
    }

    public async Task<Post?> GetPostAsync(Guid postId, string subredditHandle, string userHandle)
    {
        return await _txManager.GetPostAsync(postId, subredditHandle, userHandle);
    }

    public async Task<Post?> UpVote(Post post)
    {
        // not the best way to do that.
        post.NumberOfVotes++;
        return await _txManager.UpdatePostAsync(post, new PostUpdatedColumn[] {PostUpdatedColumn.NumberOfVotes});
    }

    public async Task<Post?> DownVote(Post post)
    {
        // not the best way to do that.
        post.NumberOfVotes--;
        return await _txManager.UpdatePostAsync(post, new PostUpdatedColumn[] {PostUpdatedColumn.NumberOfVotes});
    }

    public async Task<IEnumerable<Post>> GetPostsAsync()
    {
        return await _txManager.GetPostsAsync();
    }
}
