using DistributedReddit.AuthDb;
using DistributedReddit.Services;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using rdb_grpc;

namespace DistributedReddit.Web.Pages;
[Authorize]
public class IndexModel : PageModel
{
    private readonly ILogger<IndexModel> _logger;
    private readonly PostService _postService;
    private readonly UserManager<AuthUser> _userManager;
    private readonly UserService _userService;

    public IndexModel(ILogger<IndexModel> logger,
        PostService postService,
        UserManager<AuthUser> userManager,
        UserService userService )
    {
        _logger = logger;
        _postService = postService;
        _userManager = userManager;
        _userService = userService;
    }
    public User RdbUser { get; set; }

    public IEnumerable<Post> Posts { get; set; }

    public async Task OnGetAsync()
    {

        var authUser = await _userManager.GetUserAsync(User);
        RdbUser = await _userService.GetUserAsync(authUser.Handle);


        Posts = new List<Post>()
        {
            new Post {Title = "Hello"},
            new Post {Title = "Goodbye"},
            new Post {Title = "Yolooo"},
            new Post {Title = "Sayed"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
            new Post {Title = "Kappa"},
        };
    }
}
