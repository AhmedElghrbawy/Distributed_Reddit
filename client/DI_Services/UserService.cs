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

}
