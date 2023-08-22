# Distributed Reddit Clone

A distributed fault-tolerant Reddit clone built as an informal final project for [MIT's 6.5840 Distributed Systems](https://pdos.csail.mit.edu/6.824/).

# Application Architecture

![Slide1](https://github.com/AhmedElghrbawy/Distributed_Reddit/assets/42124024/17bc7df3-fe6e-41f7-a265-1513f52d8e05)

The application consists of:
- A [web application](https://github.com/AhmedElghrbawy/Distributed_Reddit/tree/master/client/DistributedReddit.Web) built in ASP.NET Razor Pages that communicates with a transaction manager.
- The [transaction manager](https://github.com/AhmedElghrbawy/Distributed_Reddit/tree/master/client/TransactionManager) is written as a .NET class library. It has a [gRPC client factory](https://learn.microsoft.com/en-us/aspnet/core/grpc/clientfactory) that it uses to send transaction information to a specific database replication server.
- The database layer is statically sharded using hash-based sharding where the sharding key depends on the database table.
- Each shard consists of a fixed number of replicas.
- Each replica consists of two layers, both implemented in Golang:
  1. A [replicated DB server](https://github.com/AhmedElghrbawy/Distributed_Reddit/tree/master/replicated_db/cmd/replicated_db) that accepts transaction information from the transaction manager and communicates with the database.
  2. An implementation of [Raft](https://github.com/AhmedElghrbawy/Distributed_Reddit/tree/master/replicated_db/pkg/raft), which is used by the replicated DB server to provide fault tolerance.

Note: This implementation of Raft is my own solution for [the lab](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html) but modified to fit into this project.

# Database Schema

![er_diagram drawio](https://github.com/AhmedElghrbawy/Distributed_Reddit/assets/42124024/d09924bf-6e0a-4708-9d2e-759c52a086e2)

- Posts and comments are duplicated, if needed, in the owner shard and the subreddit shard. This duplication is managed by the transaction manager.
- Joining tables (Subreddit_Users, Followage) are also duplicated if needed.
- Comments are sharded according to the post they belong to.

# Improvements

There are multiple ways to improve this project, the following are some of them:
1. Make the replicated DB server application agnostic
    - The current implementation of the replicated DB server has knowledge of the application state it is replicating. This makes it harder to add new features or modify old ones since this will require modifications in multiple layers of the application.
    - One way to solve this is to make the database return its results in a language independent format like [JSON](https://www.postgresql.org/docs/current/functions-json.html).
2. Improve Raft implementation
    - Add log compaction.
3. ...

# Screenshots

![2023-08-22 19_20_03-README md - Visual Studio Code](https://github.com/AhmedElghrbawy/Distributed_Reddit/assets/42124024/9150da25-9793-4e29-a9fa-3f0152a19b31)
![2023-08-22 19_20_56-Settings](https://github.com/AhmedElghrbawy/Distributed_Reddit/assets/42124024/74c9c4f1-9f2e-4ad5-ac3e-22ced2892bf9)
