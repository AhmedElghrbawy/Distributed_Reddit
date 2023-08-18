using System.ComponentModel.DataAnnotations;
using DistributedReddit.AuthDb;
using DistributedReddit.Services;
using Google.Protobuf;
using Google.Protobuf.WellKnownTypes;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.AspNetCore.Mvc.Rendering;
using rdb_grpc;

namespace DistributedReddit.Web.Pages.Posts;

[Authorize]

public class IndexModel : PageModel
{
    private readonly UserManager<AuthUser> _userManager;
    private readonly PostService _postService;
    private readonly UserService _userService;
    public IndexModel(
            UserManager<AuthUser> userManager,
            PostService postService,
            UserService userService
        )
    {
        _userManager = userManager;
        _postService = postService;
        _userService = userService;
    }
    public Post Post { get; set; }
    [BindProperty(SupportsGet = true)]
    public string SubredditHandle {get; set;}
    

    public  async Task<IActionResult> OnGetAsync(string id)
    {
        Post = await _postService.GetPostAsync(Guid.Parse(id), SubredditHandle);
        return Page();
    }

}

