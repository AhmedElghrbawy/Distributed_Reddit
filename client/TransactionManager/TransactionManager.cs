namespace RDB.TransactionManager;

public class TransactionManager : ITransactionManager
{
    public string SubmitTransactions(List<string> txs)
    {
        return String.Join(", ", txs);
    }
}
