using Google.Protobuf;
using Grpc.Core;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class UserTransactionManager
{
    private readonly ITransactionManager _txManager;
    private readonly ITransactionManagerConfig _config;
    private readonly GrpcClientFactory _grpcClientFactory;

    public UserTransactionManager(ITransactionManager txManager, ITransactionManagerConfig config,
     GrpcClientFactory grpcClientFactory)
    {
        _txManager = txManager;
        _config = config;
        _grpcClientFactory = grpcClientFactory;
    }


    public async Task<User?> CreateUserAsync(User user)
    {
        int shardNumber = GetUserShardNumber(user, _config.NumberOfShards);
        var clients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            clients.Add(_grpcClientFactory.CreateClient<UserGRPC.UserGRPCClient>(
                $"{nameof(UserGRPC.UserGRPCClient)}/S{shardNumber}_R{i}"
            ));
        }

        var userInfo = new UserInfo
        {
            User = user,
            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString(),
            }
        };

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var inputUserInfo = (UserInfo)inputMessage;

            return (IMessage)await userClient.CreateUserAsync(inputUserInfo, cancellationToken: cancellationToken);
        }

        var txInfo = new TransactionInfo(Guid.NewGuid(), shardNumber, clients, execFunc, userInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo});

        return result.Length == 1 ? (User) result[0] : null;
    }

    internal static int GetUserShardNumber(User user, int nShards)
    {
        return user.Handle.GetDeterministicHashCode().Mod(nShards);
    }

}
