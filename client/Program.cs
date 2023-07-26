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

using var channel = GrpcChannel.ForAddress("http://localhost:50051");

var postClient = new PostGRPC.PostGRPCClient(channel);
var twopcClient = new TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient(channel);
var userClient = new UserGRPC.UserGRPCClient(channel);
var commentClient = new CommentGRPC.CommentGRPCClient(channel);

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
        Title = "updated title",
        Content = "updated content",
        OwnerHandle = "Ahmed",
        SubredditHandle = "English",
        IsPinned = true,
        NumberOfVotes = -1,
    },
    MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
    SubredditShard = 0,
    UserShard = 0,
    TwopcInfo = new TwoPhaseCommitInfo {
        TransactionId = "bb",
    },
};

postInfo.UpdatedColumns.AddRange(new[] { PostUpdatedColumn.NumberOfVotes, PostUpdatedColumn.Title, PostUpdatedColumn.Content, PostUpdatedColumn.IsPinned});

var twopcInfo = new TwoPhaseCommitInfo {
        TransactionId = "bb",
};

var userInfo = new UserInfo {
    User = new User() {
        Handle = "kappa",
        Avatar = ByteString.CopyFrom("e#>&*m16", Encoding.Unicode),
        DisplayName = "cy@",
        CreatedAt = DateTimeOffset.UtcNow.ToTimestamp(),
        Karma = 0,
    },
    MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
};

var followage = new UserFollowage {
    FromHandle = "kappa",
    ToHandle = "Ahmed",
    FromShard = 0,
    ToShard = 1,
    MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
    TwopcInfo = new TwoPhaseCommitInfo {
        TransactionId = "bb",
    }
};

var membership = new UserSubredditMembership {
    UserHandle = "kappa",
    SubredditHandle = "Life",
    UserShard = 0,
    SubredditShard = 0,
    MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
    TwopcInfo = new TwoPhaseCommitInfo {
        TransactionId = "bb",
    }
};


var commentInfo = new CommentInfo {
    Comment = new Comment {
        Content = "bufflo",
        Id = "9c8c468c-f180-46f3-ad63-af5832e17d43",
        Image = ByteString.CopyFrom("e#>&*m16", Encoding.Unicode),
        NumberOfVotes = 0,
        OwnerHandle = "Ahmed",
        PostId = "9c8c468c-f180-46f3-ad63-af5832e17d41",
        ParentCommentId = "9c8c468c-f180-46f3-ad63-af5832e17d41",
    },
        MessageInfo = new MessageInfo {
        Id = "yeeet"
    },
    SubredditShard = 0,
    UserShard = 1,
    TwopcInfo = new TwoPhaseCommitInfo {
        TransactionId = "bb",
    }
};

// System.Console.WriteLine(subInfo);

var reply = await postClient.UpdatePostAsync(postInfo);

// var reply = await twopcClient.RollbackAsync(twopcInfo);

System.Console.WriteLine(reply);


// await Task.Delay(1000);

// var reply2 = await client.GetSubredditAsync(subInfo);

// Console.WriteLine(reply2);