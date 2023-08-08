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

    public async Task<User?> GetUserAsync(string handle)
    {
        var user  = new User{Handle = handle};
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

            return (IMessage)await userClient.GetUserAsync(inputUserInfo, cancellationToken: cancellationToken);
        }

        var txInfo = new TransactionInfo(Guid.NewGuid(), shardNumber, clients, execFunc, userInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo});

        return result.Length == 1 ? (User) result[0] : null;
    }

    public async Task<User?> UpdateUserAsync(User user, IEnumerable<UserUpdatedColumn> updatedColumns)
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

        userInfo.UpdatedColumns.AddRange(updatedColumns);

        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var inputUserInfo = (UserInfo)inputMessage;

            return (IMessage)await userClient.UpdateUserAsync(inputUserInfo, cancellationToken: cancellationToken);
        }

        var txInfo = new TransactionInfo(Guid.NewGuid(), shardNumber, clients, execFunc, userInfo);

        var result = await _txManager.SubmitTransactionsAsync(new TransactionInfo[] {txInfo});

        return result.Length == 1 ? (User) result[0] : null;
    }


    public async Task FollowAsync(string from_handle, string to_handle)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var followageInfo = (UserFollowage)inputMessage;

            return (IMessage)await userClient.FollowAsync(followageInfo, cancellationToken: cancellationToken);
        }

        await UpdateFollowageAsync(from_handle, to_handle, execFunc); 
    }

    public async Task UnFollowAsync(string from_handle, string to_handle)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var followageInfo = (UserFollowage)inputMessage;

            return (IMessage)await userClient.UnfollowAsync(followageInfo, cancellationToken: cancellationToken);
        }

        await UpdateFollowageAsync(from_handle, to_handle, execFunc); 
    }

    private async Task UpdateFollowageAsync(string from_handle, string to_handle, Func<IMessage, ClientBase, CancellationToken, Task<IMessage>> execFunc)
    {
        var txId = Guid.NewGuid();
        int fromShardNumber = GetUserShardNumber(new User{Handle = from_handle}, _config.NumberOfShards);
        int toShardNumber = GetUserShardNumber(new User{Handle = to_handle}, _config.NumberOfShards);
        var fromClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            fromClients.Add(_grpcClientFactory.CreateClient<UserGRPC.UserGRPCClient>(
                $"{nameof(UserGRPC.UserGRPCClient)}/S{fromShardNumber}_R{i}"
            ));
        }

        var toClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            toClients.Add(_grpcClientFactory.CreateClient<UserGRPC.UserGRPCClient>(
                $"{nameof(UserGRPC.UserGRPCClient)}/S{toShardNumber}_R{i}"
            ));
        }

        var followageInfo = new UserFollowage
        {
            FromHandle = from_handle,
            ToHandle = to_handle,
            FromShard = fromShardNumber,
            ToShard = toShardNumber,
            MessageInfo = new MessageInfo 
            {
                Id = Guid.NewGuid().ToString(),
            },
            TwopcInfo = new TwoPhaseCommitInfo
            {
                TransactionId = txId.ToString(),
            }
        };

        var fromTxInfo = new TransactionInfo(txId, fromShardNumber, fromClients, execFunc, followageInfo);
        var toTxInfo = new TransactionInfo(txId, toShardNumber, toClients, execFunc, followageInfo);

        var txs = new List<TransactionInfo> {fromTxInfo};

        if (fromShardNumber != toShardNumber)
        {
            txs.Add(toTxInfo);
        }
        await _txManager.SubmitTransactionsAsync(txs.ToArray());
    }

    public async Task JoinSubredditAsync(string userHandle, string subredditHandle)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var followageInfo = (UserSubredditMembership)inputMessage;

            return (IMessage)await userClient.JoinSubredditAsync(followageInfo, cancellationToken: cancellationToken);
        }

        await UpdateUserSubredditMemberShip(userHandle, subredditHandle, execFunc);
    }

    public async Task LeaveSubredditAsync(string userHandle, string subredditHandle)
    {
        static async Task<IMessage> execFunc(IMessage inputMessage, ClientBase client, CancellationToken cancellationToken)
        {
            var userClient = (UserGRPC.UserGRPCClient)client;
            var followageInfo = (UserSubredditMembership)inputMessage;

            return (IMessage)await userClient.LeaveSubredditAsync(followageInfo, cancellationToken: cancellationToken);
        }

        await UpdateUserSubredditMemberShip(userHandle, subredditHandle, execFunc);
    }

    private async Task UpdateUserSubredditMemberShip(string userHandle, string subredditHandle, Func<IMessage, ClientBase, CancellationToken, Task<IMessage>> execFunc)
    {
        var txId = Guid.NewGuid();
        int userShardNumber = GetUserShardNumber(new User{Handle = userHandle}, _config.NumberOfShards);
        int subredditShardNumber = SubredditTransactionManager.GetSubreddditShardNumber(new Subreddit{Handle = subredditHandle}, _config.NumberOfShards);

        var userClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            userClients.Add(_grpcClientFactory.CreateClient<UserGRPC.UserGRPCClient>(
                $"{nameof(UserGRPC.UserGRPCClient)}/S{userShardNumber}_R{i}"
            ));
        }

        var subredditClients = new List<ClientBase>();
        for (int i = 0; i < _config.NumberOfReplicas; i++)
        {
            subredditClients.Add(_grpcClientFactory.CreateClient<UserGRPC.UserGRPCClient>(
                $"{nameof(UserGRPC.UserGRPCClient)}/S{subredditShardNumber}_R{i}"
            ));
        }

        var userSubredditMembership = new UserSubredditMembership
        {
            UserHandle = userHandle,
            SubredditHandle = subredditHandle,
            UserShard = userShardNumber,
            SubredditShard = subredditShardNumber,

            MessageInfo = new MessageInfo
            {
                Id = Guid.NewGuid().ToString()
            },
            TwopcInfo = new TwoPhaseCommitInfo
            {
                TransactionId = txId.ToString()
            }
        };

        var userTxInfo = new TransactionInfo(txId, userShardNumber, userClients, execFunc, userSubredditMembership);
        var subredditTxInfo = new TransactionInfo(txId, subredditShardNumber, subredditClients, execFunc, userSubredditMembership);

        var txs = new List<TransactionInfo> {userTxInfo};

        if (userShardNumber != subredditShardNumber)
        {
            txs.Add(subredditTxInfo);
        }
        await _txManager.SubmitTransactionsAsync(txs.ToArray());
    }
    

    internal static int GetUserShardNumber(User user, int nShards)
    {
        return user.Handle.GetDeterministicHashCode().Mod(nShards);
    }

}
