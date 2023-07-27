namespace RDB.TransactionManager;

public interface ITransactionManager
{
    string SubmitTransactions(List<string> txs);
}
