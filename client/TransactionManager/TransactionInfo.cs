using Google.Protobuf;
using Grpc.Core;

namespace RDB.TransactionManager;

public record class TransactionInfo(List<ClientBase> Clients, Func<IMessage, ClientBase, Task<IMessage>> ExecutionFunction, IMessage InputMessage);
