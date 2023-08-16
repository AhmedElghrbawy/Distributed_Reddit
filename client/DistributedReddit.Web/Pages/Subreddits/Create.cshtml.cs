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
class CreateModel : PageModel
{
    private readonly UserManager<AuthUser> _userManager;
    private readonly SubredditService _subredditService;

    public CreateModel(
            UserManager<AuthUser> userManager,
            SubredditService subredditService
        )
    {
        _userManager = userManager;
        _subredditService = subredditService;
    }  
    [BindProperty]
    public Subreddit Subreddit { get; set; }
    [BindProperty]
    public IFormFile AvatarFile { get; set; }
    public async Task<IActionResult> OnGetAsync()
    {
        return Page();
    }

    public async Task<IActionResult> OnPostAsync()
    {
        if (!ModelState.IsValid || Subreddit == null)
        {
            return Page();
        }

        var user = await _userManager.GetUserAsync(User);


        using var ms = new MemoryStream();
        AvatarFile.CopyTo(ms);
        var AvatarBytes = ms.ToArray();

        Subreddit.Avatar = ByteString.CopyFrom(AvatarBytes);
        Subreddit.CreatedAt = Timestamp.FromDateTime(DateTime.UtcNow);
        Subreddit.AdminsHandles.Add(user.Handle);
        Subreddit.UsersHandles.Add(user.Handle);
        
        await _subredditService.CreateSubredditAsync(Subreddit);

        return RedirectToPage("Index", new {handle = Subreddit.Handle});
    }


}