
namespace RDB.TransactionManager;

public interface ITransactionManagerConfig
{
    public int NumberOfShards { get; init; }
    public int NumberOfReplicas { get; init; }
    public List<List<string>> ShardRdbServersIps { get; init; }
}
