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
public class CreateModel : PageModel
{
    private readonly UserManager<AuthUser> _userManager;
    private readonly PostService _postService;
    private readonly UserService _userService;

    public CreateModel(
            UserManager<AuthUser> userManager,
            PostService postService,
            UserService userService
        )
    {
        _userManager = userManager;
        _postService = postService;
        _userService = userService;
    }

    [BindProperty]
    public Post Post { get; set; }

    [BindProperty]
    [Display(Name="Image")]
    public IFormFile? ImageFile { get; set; }

    public IEnumerable<SelectListItem> UserSubreddits { get; set; }
    public async Task<IActionResult> OnGetAsync(string? routedSubredditHandle)
    {
        var authUser = await _userManager.GetUserAsync(User);
        var rdbUser = await _userService.GetUserAsync(authUser.Handle);

        UserSubreddits = rdbUser.AdminedSubredditHandles.Concat(rdbUser.JoinedSubredditHandles).Select(subredditHande => {
            bool isSelected = subredditHande == routedSubredditHandle;
            return new SelectListItem { Value = subredditHande, Text = $"r/{subredditHande}", Selected = isSelected };  
        });

        return Page();
    }


    public async Task<IActionResult> OnPostAsync()
    {
        if (!ModelState.IsValid || Post == null)
        {
            return Page();
        }

        var authUser = await _userManager.GetUserAsync(User);
        
        if (ImageFile is not null)
        {
            using var ms = new MemoryStream();
            ImageFile.CopyTo(ms);
            var ImageBytes = ms.ToArray();
            Post.Image = ByteString.CopyFrom(ImageBytes);
        }
        

        Post.Id = Guid.NewGuid().ToString();
        Post.CreatedAt = Timestamp.FromDateTime(DateTime.UtcNow);
        Post.OwnerHandle = authUser.Handle;
        

        await _postService.CreatePostAsync(Post);

        return RedirectToPage("/Subreddits/Index", new {handle = Post.Subreddit.Handle});
    }

}

