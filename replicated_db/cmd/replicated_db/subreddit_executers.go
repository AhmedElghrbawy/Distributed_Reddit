package main

import (
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
)

const DurationTO = 1 * time.Second

// returns (*SubredditDTO, error): the subreddit requested or error
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

// returns (bool, error): true if subreddit created successfully or error
type CreateSubredditExecuter struct {
	In_subreddit_info *pb.SubredditInfo
}

func (ex *CreateSubredditExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute CreateSubreddit for subreddit {Handle: %s}\n",
		ex.In_subreddit_info.Subreddit.Handle)
	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	isPreparedTx := ex.In_subreddit_info.SubredditShard != int32(rdb.shardNum) || ex.In_subreddit_info.UserShard != int32(rdb.shardNum)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, errors.New("couldn't start create subreddit transaction")
	}
	defer tx.Rollback()

	// insert subreddit
	subredditModel := model.Subreddits{
		Handle:    ex.In_subreddit_info.Subreddit.Handle,
		Title:     ex.In_subreddit_info.Subreddit.Title,
		About:     ex.In_subreddit_info.Subreddit.About,
		Avatar:    ex.In_subreddit_info.Subreddit.Avatar,
		Rules:     ex.In_subreddit_info.Subreddit.Rules,
		CreatedAt: ex.In_subreddit_info.Subreddit.CreatedAt.AsTime(),
	}

	subredditInsertStmt := Subreddits.INSERT(Subreddits.AllColumns).
		MODEL(subredditModel)

	_, err = subredditInsertStmt.ExecContext(ctx, tx)

	if err != nil {
		log.Printf("failed to insert subreddit {handle: %s}. rolling back tx\n", ex.In_subreddit_info.Subreddit.Handle)
		return false, err
	}

	// insert admin
	var admins []model.SubredditUsers

	for _, adminHandle := range ex.In_subreddit_info.Subreddit.AdminsHandles {
		admins = append(admins, model.SubredditUsers{
			UserHandle:      adminHandle,
			SubredditHandle: ex.In_subreddit_info.Subreddit.Handle,
			IsAdmin:         true,
		})
	}

	adminInsertStmt := SubredditUsers.INSERT(SubredditUsers.AllColumns).MODELS(admins)

	_, err = adminInsertStmt.ExecContext(ctx, tx)

	if err != nil {
		log.Printf("failed to insert admins %v for subreddit {handle: %s}. rolling back tx\n", admins, ex.In_subreddit_info.Subreddit.Handle)
		return false, err
	}

	if isPreparedTx {
		preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.In_subreddit_info.TwopcInfo.TransactionId, rdb.shardNum, rdb.replicaNum)
		preparedTxRaw := fmt.Sprintf("PREPARE TRANSACTION '%s'", preparedTxId)

		preparedStmt := RawStatement(preparedTxRaw)

		_, err = preparedStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to prepare create post tx for Subreddit {Handle: %s}.\n", ex.In_subreddit_info.Subreddit.Handle)
			return false, err
		}
		tx.Commit()
		log.Printf("successfully prepared tx for creating ubreddit {Handle: %s}\n", ex.In_subreddit_info.Subreddit.Handle)
	} else {
		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return false, errors.New("failed to commit create subreddit tx")
		}

		log.Printf("successfully commited tx for creating subreddit {Handle: %s}\n", ex.In_subreddit_info.Subreddit.Handle)

	}

	return true, nil
}

// returns (*[]SubredditDTO, error): a slice of subreddit handles on this shard or error
type GetSubredditsHandlesExecuter struct {
}

func (ex *GetSubredditsHandlesExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute GetSubreddits Handles\n")

	stmt := SELECT(
		Subreddits.Handle,
	).FROM(Subreddits)

	result := []SubredditDTO{}

	err = stmt.Query(db, &result)

	if err != nil {
		log.Printf("GetSubreddit Handles command failed %v\n", err)
		return &result, err
	}

	log.Printf("GetSubreddits Handles command completed\n")

	return &result, nil
}
