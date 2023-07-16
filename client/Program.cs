using System.Threading.Tasks;
using Google.Protobuf;
using Google.Protobuf.WellKnownTypes;
using Grpc.Net.Client;
using rdb_grpc;
using System.Text;

System.Console.WriteLine("Client starting...");

var handler = new HttpClientHandler
{
    ServerCertificateCustomValidationCallback = HttpClientHandler.DangerousAcceptAnyServerCertificateValidator
};

using var channel = GrpcChannel.ForAddress("http://localhost:50050");

var client = new SubredditGRPC.SubredditGRPCClient(channel);

// var subInfo = new SubredditInfo{
//     Subreddit = new Subreddit{
//         Handle = "Hell",
//         Title = "Titil",
//         About = "Hello this is hell",
//         Avatar = ByteString.CopyFrom("e#>&*m16", Encoding.Unicode),
//         CreatedAt = DateTimeOffset.UtcNow.ToTimestamp(),
//     },
//     MessageInfo = new MessageInfo{
//         Id = "first_message",
//     }    
// };

// subInfo.Subreddit.AdminsHandles.AddRange(new [] { "Hero", "Sayed", "paka" });

var messageInfo = new MessageInfo{
    Id = "heeeey"
};

// System.Console.WriteLine(subInfo);

var reply = await client.GetSubredditsHandlesAsync(messageInfo);

System.Console.WriteLine(reply);


// await Task.Delay(1000);

// var reply2 = await client.GetSubredditAsync(subInfo);

// Console.WriteLine(reply2);