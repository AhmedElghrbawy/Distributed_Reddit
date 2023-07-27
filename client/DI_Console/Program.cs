using DI_Services;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using RDB.TransactionManager;
using System.Text.Json;

var configuration = new ConfigurationBuilder()
    .SetBasePath(Directory.GetCurrentDirectory())
    .AddJsonFile("appsettings.json")
    .Build();

string txManagerConfPath = configuration.GetRequiredSection("tx_manager_config_file_path").Get<string>()!;

var txManagerConfig = JsonSerializer.Deserialize<TransactionManagerConfig>(File.ReadAllText(txManagerConfPath));


HostApplicationBuilder builder = Host.CreateApplicationBuilder(args);
builder.Services.AddTransient<ITransactionManager, TransactionManager>();
builder.Services.AddTransient<SubredditService>();
builder.Services.AddSingleton<ITransactionManagerConfig>(txManagerConfig);


using IHost host = builder.Build();

// Application code should start here
using IServiceScope serviceScope = host.Services.CreateScope();
IServiceProvider provider = serviceScope.ServiceProvider;

var SubredditService = provider.GetRequiredService<SubredditService>();

System.Console.WriteLine(SubredditService.CreateSubreddit());
 

await host.RunAsync();