using System.Text;
using Google.Protobuf;
using Google.Protobuf.WellKnownTypes;
using Grpc.Core;
using Grpc.Net.Client;
using Grpc.Net.ClientFactory;
using rdb_grpc;

namespace RDB.TransactionManager;

public class TransactionManager : ITransactionManager
{
    private readonly ITransactionManagerConfig _config;

    public TransactionManager(ITransactionManagerConfig config)
    {
        _config = config;
    }
    public async Task<List<IMessage>> SubmitTransactionsAsync(List<TransactionInfo> txs)
    {       
        var result = new List<IMessage>();
        if (txs.Count == 1)
        {
            result.Add(await RunTransactionAsync(txs[0]));
            return result;
        }      

        throw new NotImplementedException(); 
    }

    private async Task<IMessage> RunTransactionAsync(TransactionInfo txInfo)
    {
        // TODO: add a timout duration around this method.

        for (int i = 0; i < txInfo.Clients.Count; i = (i + 1) % _config.NumberOfReplicas)
        {
            var client = txInfo.Clients[i];
            try 
            {
                return await txInfo.ExecutionFunction(txInfo.InputMessage, client);
            }
            catch (RpcException ex) 
                when (ex.StatusCode == StatusCode.Unavailable || ex.StatusCode == StatusCode.DeadlineExceeded)
            {
                // retry with different client
                System.Console.WriteLine(ex.Message);
                continue;
            }
        }        
        throw new TimeoutException(); 
    }

}
