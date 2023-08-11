using System.Text.Json;
using DistributedReddit.AuthDb;
using DistributedReddit.Services;
using Microsoft.EntityFrameworkCore;
using RDB.TransactionManager;
using rdb_grpc;

var builder = WebApplication.CreateBuilder(args);

string txManagerConfPath = builder.Configuration.GetRequiredSection("tx_manager_config_file_path").Get<string>()!;
var txManagerConfig = JsonSerializer.Deserialize<TransactionManagerConfig>(File.ReadAllText(txManagerConfPath))!;

var authDbConnectionString = builder.Configuration["DistributedReddit:authDbConnectionStr"];


// Add services to the container.
builder.Services.AddRazorPages();

builder.Services.AddSingleton<ITransactionManager, TransactionManager>();
builder.Services.AddSingleton<ITransactionManagerConfig>(txManagerConfig);

builder.Services.AddTransient<SubredditService>();
builder.Services.AddTransient<PostService>();
builder.Services.AddTransient<UserService>();
builder.Services.AddTransient<CommentService>();

builder.Services.AddTransient<SubredditTransactionManager>();
builder.Services.AddTransient<PostTransactionManager>();
builder.Services.AddTransient<UserTransactionManager>();
builder.Services.AddTransient<CommentTransactionManager>();

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


builder.Services.AddDbContext<AuthDbContext>(options => 
{
    options.UseNpgsql(authDbConnectionString);
});

builder.Services.AddDefaultIdentity<AuthUser>(options => options.SignIn.RequireConfirmedAccount = false).AddEntityFrameworkStores<AuthDbContext>();

var app = builder.Build();

// Configure the HTTP request pipeline.
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Home/Error");
    // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
    app.UseHsts();
}

app.UseHttpsRedirection();
app.UseStaticFiles();

app.UseRouting();

app.UseAuthorization();

app.MapRazorPages();


app.Run();
