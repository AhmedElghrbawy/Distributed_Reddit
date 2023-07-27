namespace RDB.TransactionManager;

public class TransactionManager : ITransactionManager
{
    private readonly ITransactionManagerConfig _config;

    public TransactionManager(ITransactionManagerConfig config)
    {
        _config = config;
    }
    public string SubmitTransactions(List<string> txs)
    {
        return String.Join(", ", txs);
    }
}
