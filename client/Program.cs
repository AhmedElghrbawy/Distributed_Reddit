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

using var channel = GrpcChannel.ForAddress("http://localhost:50052");

var client = new PostGRPC.PostGRPCClient(channel);

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

var postInfo = new PostInfo{
    Post = new Post {
        Id = "9c8c468c-f180-46f3-ad63-af5832e17d41",
    },
    MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
    SubredditShard = 0,
    UserShard = 1,
};

// System.Console.WriteLine(subInfo);

var reply = await client.GetPostAsync(postInfo);

System.Console.WriteLine(reply);


// await Task.Delay(1000);

// var reply2 = await client.GetSubredditAsync(subInfo);

// Console.WriteLine(reply2);