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
    private readonly UserService _userService;

    public IndexModel(
            UserManager<AuthUser> userManager,
            SubredditService subredditService,
            UserService userService
        )
    {
        _userManager = userManager;
        _subredditService = subredditService;
        _userService = userService;
    }

    [BindProperty]
    public Subreddit Subreddit { get; set; }
    public User RdbUser { get; set; }



    public async Task<IActionResult> OnGetAsync(string handle)
    {
        Subreddit = await _subredditService.GetSubredditAsync(handle);
        var authUser = await _userManager.GetUserAsync(User);
        RdbUser = await _userService.GetUserAsync(authUser.Handle);

        if (Subreddit == null)
            return NotFound();

        return Page();
    }

    public async Task<IActionResult> OnPostLeaveAsync()
    {
        if (!ModelState.IsValid)
        {
            return Page();
        }

        var authUser = await _userManager.GetUserAsync(User);
        
        await _userService.LeaveSubredditAsync(authUser.Handle, Subreddit.Handle);

        return RedirectToPage("Index");
    }
    public async Task<IActionResult> OnPostJoinAsync()
    {
        if (!ModelState.IsValid)
        {
            return Page();
        }

        var authUser = await _userManager.GetUserAsync(User);
        
        await _userService.JoinSubredditAsync(authUser.Handle, Subreddit.Handle);

        return RedirectToPage("Index");
    }
}
