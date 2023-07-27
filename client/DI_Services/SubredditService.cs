using RDB.TransactionManager;

namespace DI_Services;

public class SubredditService
{
    private readonly ITransactionManager _tx;

    public SubredditService(ITransactionManager tx)
    {
        _tx = tx;
    }

    public string CreateSubreddit()
    {
        return _tx.SubmitTransactions(new List<string> {"Hey", "you", "there"});
    }
}
