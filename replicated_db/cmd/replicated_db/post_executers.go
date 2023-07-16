package main

import (
	"database/sql"
	"log"

	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
)

// returns (*PostDTO, error): the post requested or error
type GetPostExecuter struct {
	In_post_info *pb.PostInfo
}

func (ex *GetPostExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute GetPost for post {Id: %s}\n",
		ex.In_post_info.Post.Id)

	var stmt SelectStatement
	if ex.In_post_info.UserShard == int32(rdb.shardNum) {
		stmt = SELECT(
			UserPosts.AllColumns, UserComments.AllColumns,
			PostTags.TagName,
		).FROM(UserPosts.
			LEFT_JOIN(UserComments, UserPosts.ID.EQ(UserComments.PostID)).
			LEFT_JOIN(PostTags, UserPosts.ID.EQ(PostTags.PostID)),
		).WHERE(
			CAST(table.UserPosts.ID).AS_TEXT().EQ(String(ex.In_post_info.Post.Id)),
		)
	} else {
		stmt = SELECT(
			SubredditPosts.AllColumns, SubredditComments.AllColumns,
			PostTags.TagName,
		).FROM(SubredditPosts.
			LEFT_JOIN(SubredditComments, SubredditPosts.ID.EQ(SubredditComments.PostID)).
			LEFT_JOIN(PostTags, SubredditPosts.ID.EQ(PostTags.PostID)),
		).WHERE(
			CAST(table.SubredditPosts.ID).AS_TEXT().EQ(String(ex.In_post_info.Post.Id)),
		)

	}

	result := PostDTO{}
	err = stmt.Query(db, &result)

	if err != nil {
		log.Printf("GetPost {Id: %s} command failed %v\n", ex.In_post_info.Post.Id, err)
		return &PostDTO{}, err
	}

	log.Printf("GetPost {Id: %s} command completed\n", ex.In_post_info.Post.Id)

	return &result, nil
}
