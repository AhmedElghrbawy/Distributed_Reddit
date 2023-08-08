using System.Diagnostics;
using Microsoft.AspNetCore.Mvc;
using DistributedReddit.Web.Models;
using DistributedReddit.Services;

namespace DistributedReddit.Web.Controllers;

public class HomeController : Controller
{
    private readonly ILogger<HomeController> _logger;
    private readonly PostService _postService;

    public HomeController(ILogger<HomeController> logger, PostService postService)
    {
        _logger = logger;
        _postService = postService;
    }

    public async Task<IActionResult> Index()
    {
        ViewData["posts"] = await _postService.GetPostsAsync();
        return View();
    }

    public IActionResult Privacy()
    {
        return View();
    }

    [ResponseCache(Duration = 0, Location = ResponseCacheLocation.None, NoStore = true)]
    public IActionResult Error()
    {
        return View(new ErrorViewModel { RequestId = Activity.Current?.Id ?? HttpContext.TraceIdentifier });
    }
}
