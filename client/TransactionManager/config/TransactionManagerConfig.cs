using System.Text.Json.Serialization;

namespace RDB.TransactionManager;

public record TransactionManagerConfig : ITransactionManagerConfig
{
    [JsonPropertyName("number_of_shards")]
    public int NumberOfShards { get; init; }
    [JsonPropertyName("number_of_replicas")]
    public int NumberOfReplicas { get; init; }
    [JsonPropertyName("shard_rdb_servers_ips")]
    public List<List<string>> ShardRdbServersIps { get; init; }
}
