using System.ComponentModel.DataAnnotations;
using Microsoft.AspNetCore.Identity;

namespace DistributedReddit.AuthDb;

public class AuthUser : IdentityUser
{
    [Required]
    // TODO: make this unique
    public string Handle { get; set; }

}
