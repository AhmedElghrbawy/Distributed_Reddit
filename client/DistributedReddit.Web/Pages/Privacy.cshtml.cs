using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace DistributedReddit.Web.Pages;

public class PrivacyModel : PageModel
{
    private readonly ILogger<IndexModel> _logger;

    public PrivacyModel(ILogger<IndexModel> logger)
    {
        _logger = logger;
    }

    public void OnGet()
    {

    }
}
