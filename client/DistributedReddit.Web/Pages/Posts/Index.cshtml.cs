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
    private readonly CommentService _commentService;
    public IndexModel(
            UserManager<AuthUser> userManager,
            PostService postService,
            UserService userService,
            CommentService commentService
        )
    {
        _userManager = userManager;
        _postService = postService;
        _userService = userService;
        _commentService = commentService;
    }
    [BindProperty]
    public Post Post { get; set; }

    [BindProperty(SupportsGet = true)]
    public string SubredditHandle {get; set;}
    

    [BindProperty]
    public Comment AddedComment { get; set; }
    public  async Task<IActionResult> OnGetAsync(string id)
    {
        System.Console.WriteLine(id);
        Post = await _postService.GetPostAsync(Guid.Parse(id), SubredditHandle);
        return Page();
    }

    public async Task<IActionResult> OnPostAddCommentAsync()
    {
        var authUser = await _userManager.GetUserAsync(User);
        
        AddedComment.PostId = Post.Id;
        AddedComment.Id = Guid.NewGuid().ToString();
        AddedComment.OwnerHandle = authUser.Handle;
        
        await _commentService.CreateCommentAsync(AddedComment, Post.Subreddit.Handle);
        System.Console.WriteLine("Redirected id" + Post.Id);
        return RedirectToPage("./Index", new {id = Post.Id, SubredditHandle = Post.Subreddit.Handle});
    }

}

