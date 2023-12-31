package main

import (
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

var grpc_column_to_subpost_column = map[pb.PostUpdatedColumn]Column{
	pb.PostUpdatedColumn_TITLE:           SubredditPosts.Title,
	pb.PostUpdatedColumn_CONTENT:         SubredditPosts.Content,
	pb.PostUpdatedColumn_IS_PINNED:       SubredditPosts.IsPinned,
	pb.PostUpdatedColumn_NUMBER_OF_VOTES: SubredditPosts.NumberOfVotes,
}

var grpc_column_to_userpost_column = map[pb.PostUpdatedColumn]Column{
	pb.PostUpdatedColumn_TITLE:           AliasedUserPosts.Title,
	pb.PostUpdatedColumn_CONTENT:         AliasedUserPosts.Content,
	pb.PostUpdatedColumn_IS_PINNED:       AliasedUserPosts.IsPinned,
	pb.PostUpdatedColumn_NUMBER_OF_VOTES: AliasedUserPosts.NumberOfVotes,
}

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
			AliasedUserPosts.AllColumns, AliasedUserComments.AllColumns,
			PostTags.TagName,
		).FROM(AliasedUserPosts.
			LEFT_JOIN(AliasedUserComments, AliasedUserPosts.ID.EQ(AliasedUserComments.PostID)).
			LEFT_JOIN(PostTags, AliasedUserPosts.ID.EQ(PostTags.PostID)),
		).WHERE(
			CAST(AliasedUserPosts.ID).AS_TEXT().EQ(String(ex.In_post_info.Post.Id)),
		)
	} else {
		stmt = SELECT(
			SubredditPosts.AllColumns, SubredditComments.AllColumns,
			PostTags.TagName, Subreddits.Avatar,
		).FROM(SubredditPosts.
			LEFT_JOIN(SubredditComments, SubredditPosts.ID.EQ(SubredditComments.PostID)).
			LEFT_JOIN(PostTags, SubredditPosts.ID.EQ(PostTags.PostID)).
			LEFT_JOIN(Subreddits, SubredditPosts.SubredditHandle.EQ(Subreddits.Handle)),
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

// returns (bool, error): true if post created/prepared successfully or error
type CreatePostExecuter struct {
	In_post_info *pb.PostInfo
}

func (ex *CreatePostExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute CreatePost for post {Id: %s}\n",
		ex.In_post_info.Post.Id)

	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	subredditPostModel := model.SubredditPosts{
		ID:              uuid.MustParse(ex.In_post_info.Post.Id),
		Title:           ex.In_post_info.Post.Title,
		Content:         ex.In_post_info.Post.Content,
		Image:           &ex.In_post_info.Post.Image,
		CreatedAt:       ex.In_post_info.Post.CreatedAt.AsTime(),
		NumberOfVotes:   ex.In_post_info.Post.NumberOfVotes,
		IsPinned:        ex.In_post_info.Post.IsPinned,
		OwnerHandle:     ex.In_post_info.Post.OwnerHandle,
		SubredditHandle: ex.In_post_info.Post.Subreddit.Handle,
	}

	subredditPostInsertStmt := SubredditPosts.INSERT(SubredditPosts.AllColumns).
		MODEL(subredditPostModel)

	userPostModels := model.UserPosts{
		ID:              uuid.MustParse(ex.In_post_info.Post.Id),
		Title:           ex.In_post_info.Post.Title,
		Content:         ex.In_post_info.Post.Content,
		Image:           &ex.In_post_info.Post.Image,
		CreatedAt:       ex.In_post_info.Post.CreatedAt.AsTime(),
		NumberOfVotes:   ex.In_post_info.Post.NumberOfVotes,
		IsPinned:        ex.In_post_info.Post.IsPinned,
		OwnerHandle:     ex.In_post_info.Post.OwnerHandle,
		SubredditHandle: ex.In_post_info.Post.Subreddit.Handle,
	}

	userPostInsertStmt := UserPosts.INSERT(UserPosts.AllColumns).
		MODEL(userPostModels)

	postTags := make([]model.PostTags, 0)

	for _, tag := range ex.In_post_info.Post.Tags {
		postTags = append(postTags, model.PostTags{
			TagName: tag,
			PostID:  uuid.MustParse(ex.In_post_info.Post.Id),
		})
	}

	tagsInsertStmt := PostTags.INSERT(PostTags.AllColumns).MODELS(postTags)

	isPreparedTx := ex.In_post_info.UserShard != int32(rdb.shardNum) || ex.In_post_info.SubredditShard != int32(rdb.shardNum)
	isUserShard := ex.In_post_info.UserShard == int32(rdb.shardNum)
	isSubredditShard := ex.In_post_info.SubredditShard == int32(rdb.shardNum)

	if !isPreparedTx {
		log.Printf("Creating a new post {Id: %s} as part of a non-prepared transaction", ex.In_post_info.Post.Id)
	} else if isUserShard {
		log.Printf("Creating a new userPost {Id: %s} as part of a prepared transaction", ex.In_post_info.Post.Id)
	} else {
		log.Printf("Creating a new subredditPost {Id: %s} as part of a prepared transaction", ex.In_post_info.Post.Id)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, errors.New("couldn't start create post transaction")
	}
	defer tx.Rollback()

	if isSubredditShard {
		// create subreddit post
		_, err = subredditPostInsertStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to insert subredditPost {Id: %s}. rolling back tx\n", ex.In_post_info.Post.Id)
			return false, err
		}
	}

	if isUserShard {
		// create user post
		_, err = userPostInsertStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to insert userPost {Id: %s}. rolling back tx\n", ex.In_post_info.Post.Id)
			return false, err
		}
	}

	if len(postTags) > 0 {
		// create post tags
		_, err = tagsInsertStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to insert Tags: %v for Post {Id: %s}. rolling back tx\n", postTags, ex.In_post_info.Post.Id)
			return false, err
		}
	}

	if isPreparedTx {
		// prepare the transacion
		// looks like Postgres doesn't support arguments in prepare transacion statment
		// so we are going to use go Sprintf
		preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.In_post_info.TwopcInfo.TransactionId, rdb.shardNum, rdb.replicaNum)
		preparedTxRaw := fmt.Sprintf("PREPARE TRANSACTION '%s'", preparedTxId)

		preparedStmt := RawStatement(preparedTxRaw)

		_, err = preparedStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to prepare create post tx for Post {Id: %s}.\n", ex.In_post_info.Post.Id)
			return false, err
		}

		// we are not actually commiting the tx here
		// this is a workaround since Sql.Tx doesn't support prepared transactions
		// this is only done to free the connection
		// https://github.com/go-pg/pg/issues/490
		tx.Commit()
		log.Printf("successfully prepared tx for creating post {Id: %s}\n", ex.In_post_info.Post.Id)

	} else {
		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return false, errors.New("failed to commit create post tx")
		}
		log.Printf("successfully commited tx for creating post {Id: %s}\n", ex.In_post_info.Post.Id)
	}

	return true, nil

}

// returns (*[]PostDTO, error): a slice of subreddit post with tags on this shard or error
type GetPostsExecuter struct {
}

func (ex *GetPostsExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute GetPosts\n")

	stmt := SELECT(
		SubredditPosts.AllColumns, SubredditComments.AllColumns,
		PostTags.TagName, Subreddits.Avatar,
	).FROM(SubredditPosts.
		LEFT_JOIN(SubredditComments, SubredditPosts.ID.EQ(SubredditComments.PostID)).
		LEFT_JOIN(PostTags, SubredditPosts.ID.EQ(PostTags.PostID)).
		LEFT_JOIN(Subreddits, SubredditPosts.SubredditHandle.EQ(Subreddits.Handle)),
	)

	result := []PostDTO{}

	err = stmt.Query(db, &result)

	if err != nil {
		log.Printf("getPosts command failed %v\n", err)
		return &result, err
	}

	log.Printf("getPosts command completed\n")

	return &result, nil
}

// returns (postDTO, error): the post after updating or error
type UpdatePostExecuter struct {
	In_post_info         *pb.PostInfo
	Updated_column_value map[pb.PostUpdatedColumn]interface{}
}

func (ex *UpdatePostExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	ex.Updated_column_value = map[pb.PostUpdatedColumn]interface{}{
		pb.PostUpdatedColumn_TITLE:           ex.In_post_info.Post.Title,
		pb.PostUpdatedColumn_CONTENT:         ex.In_post_info.Post.Content,
		pb.PostUpdatedColumn_IS_PINNED:       ex.In_post_info.Post.IsPinned,
		pb.PostUpdatedColumn_NUMBER_OF_VOTES: Int(int64(ex.In_post_info.Post.NumberOfVotes)),
	}

	subpost_columns_to_update := ColumnList{}
	for _, c := range ex.In_post_info.UpdatedColumns {
		subpost_columns_to_update = append(subpost_columns_to_update, grpc_column_to_subpost_column[c])
	}

	userpost_columns_to_update := ColumnList{}
	for _, c := range ex.In_post_info.UpdatedColumns {
		userpost_columns_to_update = append(userpost_columns_to_update, grpc_column_to_userpost_column[c])
	}

	subpost_updated_values := make([]interface{}, 0)

	for _, c := range ex.In_post_info.UpdatedColumns {
		if c == pb.PostUpdatedColumn_NUMBER_OF_VOTES {
			value := ex.Updated_column_value[c].(IntegerExpression).ADD(SubredditPosts.NumberOfVotes)
			subpost_updated_values = append(subpost_updated_values, value)
		} else {
			subpost_updated_values = append(subpost_updated_values, ex.Updated_column_value[c])
		}
	}

	userpost_updated_values := make([]interface{}, 0)

	for _, c := range ex.In_post_info.UpdatedColumns {
		if c == pb.PostUpdatedColumn_NUMBER_OF_VOTES {
			value := ex.Updated_column_value[c].(IntegerExpression).ADD(AliasedUserPosts.NumberOfVotes)
			userpost_updated_values = append(userpost_updated_values, value)
		} else {
			userpost_updated_values = append(userpost_updated_values, ex.Updated_column_value[c])
		}
	}

	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to update post {Id: %s} columns: %v, values: %v\n", ex.In_post_info.Post.Id, userpost_columns_to_update, userpost_updated_values)

	isUserPost := ex.In_post_info.UserShard == int32(rdb.shardNum)
	isSubredditPost := ex.In_post_info.SubredditShard == int32(rdb.shardNum)

	subPostStmt := SubredditPosts.
		UPDATE(subpost_columns_to_update).
		SET(subpost_updated_values[0], subpost_updated_values[1:]...).
		WHERE(CAST(table.SubredditPosts.ID).AS_TEXT().EQ(String(ex.In_post_info.Post.Id))).
		RETURNING(SubredditPosts.AllColumns)

	userPostStmt := AliasedUserPosts.
		UPDATE(userpost_columns_to_update).
		SET(userpost_updated_values[0], userpost_updated_values[1:]...).
		WHERE(CAST(AliasedUserPosts.ID).AS_TEXT().EQ(String(ex.In_post_info.Post.Id))).
		RETURNING(AliasedUserPosts.AllColumns)

	result := PostDTO{}

	if isUserPost {
		err = userPostStmt.Query(db, &result)

		if err != nil {
			log.Printf("update post {Id: %s} command failed %v\n", ex.In_post_info.Post.Id, err)
			return 0, err
		}
	}

	if isSubredditPost {
		err = subPostStmt.Query(db, &result)

		if err != nil {
			log.Printf("update post {Id: %s} command failed %v\n", ex.In_post_info.Post.Id, err)
			return 0, err
		}
	}
	log.Printf("change vote value for command completed\n")

	return &result, nil
}
