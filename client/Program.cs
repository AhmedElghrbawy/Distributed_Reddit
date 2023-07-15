using System.Threading.Tasks;
using Google.Protobuf.WellKnownTypes;
using Grpc.Net.Client;
using rdb_grpc;

var handler = new HttpClientHandler
{
    ServerCertificateCustomValidationCallback = HttpClientHandler.DangerousAcceptAnyServerCertificateValidator
};

using var channel = GrpcChannel.ForAddress("http://localhost:50050");

var client = new SubredditGRPC.SubredditGRPCClient(channel);

var subInfo = new SubredditInfo{
    Subreddit = new Subreddit{
        Handle = "English"
    },
    MessageInfo = new MessageInfo{
        Id = "first_message",
    }    
};



// System.Console.WriteLine(subInfo);

var reply = await client.GetSubredditAsync(subInfo);

System.Console.WriteLine(reply);


await Task.Delay(1000);

var reply2 = await client.GetSubredditAsync(subInfo);

Console.WriteLine(reply2);