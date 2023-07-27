using DI_Services;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using RDB.TransactionManager;

HostApplicationBuilder builder = Host.CreateApplicationBuilder(args);
builder.Services.AddTransient<ITransactionManager, TransactionManager>();
builder.Services.AddTransient<SubredditService>();
using IHost host = builder.Build();





using IServiceScope serviceScope = host.Services.CreateScope();
IServiceProvider provider = serviceScope.ServiceProvider;

var SubredditService = provider.GetRequiredService<SubredditService>();

System.Console.WriteLine(SubredditService.CreateSubreddit());
 

await host.RunAsync();