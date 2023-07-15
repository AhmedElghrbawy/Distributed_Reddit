package main

import (
	"database/sql"
	"log"

	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
)

type GetSubredditExecuter struct {
	In_subreddit_info *pb.SubredditInfo
}

func (ex *GetSubredditExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute GetSubreddit for subreddit {Handle: %s}\n",
	ex.In_subreddit_info.Subreddit.Handle)

	stmt := SELECT(
		Subreddits.AllColumns, SubredditPosts.AllColumns, SubredditComments.AllColumns,
		PostTags.TagName, SubredditUsers.UserHandle,
	).FROM(Subreddits.
		LEFT_JOIN(SubredditPosts, Subreddits.Handle.EQ(SubredditPosts.SubredditHandle)).
		LEFT_JOIN(SubredditComments, SubredditPosts.ID.EQ(SubredditComments.PostID)).
		LEFT_JOIN(PostTags, SubredditPosts.ID.EQ(PostTags.PostID)).
		LEFT_JOIN(SubredditUsers, Subreddits.Handle.EQ(SubredditUsers.SubredditHandle)),
	).WHERE(
		Subreddits.Handle.EQ(String(ex.In_subreddit_info.Subreddit.Handle)),
	)

	// log.Println(stmt.DebugSql())

	result := SubredditDTO{}

	
	err = stmt.Query(db, &result)
	
	
	if err != nil {
		log.Printf("GetSubreddit {Handle: %s} command failed %v\n", ex.In_subreddit_info.Subreddit.Handle, err)
		return &SubredditDTO{}, err
	}

	log.Printf("GetSubreddit {Handle: %s} command completed\n", ex.In_subreddit_info.Subreddit.Handle)
	
	return &result, nil
}
