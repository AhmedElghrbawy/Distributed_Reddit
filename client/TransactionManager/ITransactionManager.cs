using Google.Protobuf;

namespace RDB.TransactionManager;

public interface ITransactionManager
{
    Task<List<IMessage>> SubmitTransactionsAsync(List<TransactionInfo> txs);
}
