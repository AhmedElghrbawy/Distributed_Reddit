using DistributedReddit.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using rdb_grpc;

namespace DistributedReddit.Web.Pages;

public class IndexModel : PageModel
{
    private readonly ILogger<IndexModel> _logger;
    private readonly PostService _postService;

    public IndexModel(ILogger<IndexModel> logger, PostService postService)
    {
        _logger = logger;
        _postService = postService;
    }

    public IEnumerable<Post> Posts { get; set; }

    public async Task OnGetAsync()
    {
        Posts = await _postService.GetPostsAsync();
    }
}
