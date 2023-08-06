using Google.Protobuf;
using Grpc.Core;

namespace RDB.TransactionManager;

public record class TransactionInfo(
        Guid TransactionId,
        int ShardNumber,
        List<ClientBase> Clients,
        Func<IMessage, ClientBase, CancellationToken, Task<IMessage>> ExecutionFunction, 
        IMessage InputMessage
    );
