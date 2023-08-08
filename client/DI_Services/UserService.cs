using RDB.TransactionManager;
using rdb_grpc;

namespace DI.Services;

public class UserService
{
    private readonly UserTransactionManager _txManager;

    public UserService(UserTransactionManager txManager)
    {
        _txManager = txManager;
    }

    public async Task<User?> CreateUserAsync(User user)
    {
        return await _txManager.CreateUserAsync(user);
    } 

    public async Task<User?> GetUserAsync(string handle)
    {
        return await _txManager.GetUserAsync(handle);
    }

    public async Task<User?> ChangeDisplayNameAsync(string userHandle, string displayName)
    {
        return await _txManager.UpdateUserAsync(new User {Handle = userHandle, DisplayName = displayName},
        new UserUpdatedColumn[] {UserUpdatedColumn.DisplayName});
    }

    public async Task<User?> IncreaseKarmaAsync(User user)
    {
        // not the best way to do this.
        user.Karma++;
        return await _txManager.UpdateUserAsync(user, new UserUpdatedColumn[] {UserUpdatedColumn.Karma});
    }

    
    public async Task<User?> DecreaseKarmaAsync(User user)
    {
        // not the best way to do this.
        user.Karma--;
        return await _txManager.UpdateUserAsync(user, new UserUpdatedColumn[] {UserUpdatedColumn.Karma});
    }
}
