using System.ComponentModel.DataAnnotations;
using Microsoft.AspNetCore.Identity;

namespace DistributedReddit.AuthDb;

public class AuthUser : IdentityUser
{
    [Required]
    public string Handle { get; set; }

}
