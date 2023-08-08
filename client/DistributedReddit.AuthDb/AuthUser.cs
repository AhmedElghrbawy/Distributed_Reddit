using System.ComponentModel.DataAnnotations;
using Microsoft.AspNetCore.Identity;

namespace DistributedReddit.AuthDb;

public class AuthUser : IdentityUser
{
    [Required]
    public string Handle { get; set; }

    [Required]
    public string DisplayName { get; set; }
    [Required]
    public byte[] Avatar { get; set; }
}
