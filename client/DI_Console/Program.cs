using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using RDB.TransactionManager;
using System.Text.Json;
using rdb_grpc;
using Google.Protobuf;
using System.Text;
using Google.Protobuf.WellKnownTypes;
using DI.Services;

var configuration = new ConfigurationBuilder()
    .SetBasePath(Directory.GetCurrentDirectory())
    .AddJsonFile("appsettings.json")
    .Build();

string txManagerConfPath = configuration.GetRequiredSection("tx_manager_config_file_path").Get<string>()!;

var txManagerConfig = JsonSerializer.Deserialize<TransactionManagerConfig>(File.ReadAllText(txManagerConfPath));


HostApplicationBuilder builder = Host.CreateApplicationBuilder(args);
builder.Services.AddSingleton<ITransactionManager, TransactionManager>();
builder.Services.AddTransient<SubredditService>();
builder.Services.AddTransient<PostService>();
builder.Services.AddTransient<UserService>();
builder.Services.AddTransient<SubredditTransactionManager>();
builder.Services.AddTransient<PostTransactionManager>();
builder.Services.AddTransient<UserTransactionManager>();
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

var postService = provider.GetRequiredService<UserService>();

var user = new User
{
    Handle = "gar",
    DisplayName = "Heloo",
};

System.Console.WriteLine(await postService.CreateUserAsync(user));
 

await host.RunAsync();