﻿using Google.Protobuf;
using Grpc.Core;

namespace RDB.TransactionManager;

public record class TransactionInfo(
        List<ClientBase> Clients,
        Func<IMessage, ClientBase, CancellationToken, Task<IMessage>> ExecutionFunction, 
        IMessage InputMessage
    );
