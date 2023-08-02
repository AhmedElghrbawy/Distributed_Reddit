using DI_Services;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using RDB.TransactionManager;
using System.Text.Json;
using rdb_grpc;
using Google.Protobuf;
using System.Text;
using Google.Protobuf.WellKnownTypes;

var configuration = new ConfigurationBuilder()
    .SetBasePath(Directory.GetCurrentDirectory())
    .AddJsonFile("appsettings.json")
    .Build();

string txManagerConfPath = configuration.GetRequiredSection("tx_manager_config_file_path").Get<string>()!;

var txManagerConfig = JsonSerializer.Deserialize<TransactionManagerConfig>(File.ReadAllText(txManagerConfPath));


HostApplicationBuilder builder = Host.CreateApplicationBuilder(args);
builder.Services.AddTransient<ITransactionManager, TransactionManager>();
builder.Services.AddTransient<SubredditService>();
builder.Services.AddTransient<SubredditTransactionManager>();
builder.Services.AddSingleton<ITransactionManagerConfig>(txManagerConfig);

for (int i = 0; i < txManagerConfig.NumberOfShards; i++)
{
    for (int j = 0; j < txManagerConfig.NumberOfReplicas; j++)
    {
        // ! make sure not to pass any interpolated string to options.Address. 
        // ! It seems like options is lazily evaluated (maybe because it is an Action) so it might cause a problem.
        // ! define a variable that is immediately evaluted and pass it to options.  
        var channelAddres = new Uri($"http://{txManagerConfig.ShardRdbServersIps[i][j]}");

        builder.Services.AddGrpcClient<SubredditGRPC.SubredditGRPCClient>($"{nameof(SubredditGRPC.SubredditGRPCClient)}/S{i}_R{j}", o => {
            o.Address = channelAddres;
        });
        builder.Services.AddGrpcClient<UserGRPC.UserGRPCClient>($"{nameof(UserGRPC.UserGRPCClient)}/S{i}_R{j}", o => {
            o.Address = channelAddres;
        });
        builder.Services.AddGrpcClient<PostGRPC.PostGRPCClient>($"{nameof(PostGRPC.PostGRPCClient)}/S{i}_R{j}", o => {
            o.Address = channelAddres;
        });
        builder.Services.AddGrpcClient<CommentGRPC.CommentGRPCClient>($"{nameof(CommentGRPC.CommentGRPCClient)}/S{i}_R{j}", o => {
            o.Address = channelAddres;
        });
        builder.Services.AddGrpcClient<TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient>($"{nameof(TwoPhaseCommitGRPC.TwoPhaseCommitGRPCClient)}/S{i}_R{j}", o => {
            o.Address = channelAddres;
        });
    }
}


using IHost host = builder.Build();

// Application code should start here
using IServiceScope serviceScope = host.Services.CreateScope();
IServiceProvider provider = serviceScope.ServiceProvider;

var SubredditService = provider.GetRequiredService<SubredditService>();

var subreddit = new Subreddit{
        Handle = "623423",
        Title = "ff15",
        About = "Hello this is hell",
        Avatar = ByteString.CopyFrom("e#>&*m16", Encoding.Unicode),
        CreatedAt = DateTimeOffset.UtcNow.ToTimestamp(),
};

subreddit.AdminsHandles.Add(new string[] { "Ahmed" });

System.Console.WriteLine(await SubredditService.CreateSubredditAsync(subreddit));
 

await host.RunAsync();