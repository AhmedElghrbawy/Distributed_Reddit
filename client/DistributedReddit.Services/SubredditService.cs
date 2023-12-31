﻿using RDB.TransactionManager;
using rdb_grpc;

namespace DistributedReddit.Services;

public class SubredditService
{
    private readonly SubredditTransactionManager _txManager;

    public SubredditService(SubredditTransactionManager subredditTxManager)
    {
        _txManager = subredditTxManager;
    }

    public async Task<Subreddit?> CreateSubredditAsync(Subreddit subreddit)
    {
        return await _txManager.CreateSubredditAsync(subreddit);
    }

    public async Task<Subreddit?> GetSubredditAsync(string handle)
    {
        return await _txManager.GetSubredditAsync(handle);
    }

    public async Task<IEnumerable<string>> GetSubredditsHandlesAsync()
    {
        return await _txManager.GetSubredditsHandlesAsync();
    }
}
