using DistributedReddit.AuthDb;
using DistributedReddit.Services;
using Google.Protobuf;
using Google.Protobuf.WellKnownTypes;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using rdb_grpc;


namespace DistributedReddit.Web.Pages.Subreddits;


[Authorize]
public class IndexModel : PageModel
{
    private readonly UserManager<AuthUser> _userManager;
    private readonly SubredditService _subredditService;
    public IndexModel(
            UserManager<AuthUser> userManager,
            SubredditService subredditService
        )
    {
        _userManager = userManager;
        _subredditService = subredditService;
    }


    public Subreddit Subreddit { get; set; }



    public async Task<IActionResult> OnGetAsync(string handle)
    {
        Subreddit = await _subredditService.GetSubredditAsync(handle);

        if (Subreddit == null)
            return NotFound();

        return Page();
    }
}
