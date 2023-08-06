using Google.Protobuf;

namespace RDB.TransactionManager;

public interface ITransactionManager
{
    Task<IMessage[]> SubmitTransactionsAsync(TransactionInfo[] txs);
}
