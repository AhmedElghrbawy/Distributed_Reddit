package main

import (
	context "context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type AddCommentExecuter struct {
	In_comment_info *pb.CommentInfo
}

// returns (bool, error): true if comment created/prepared successfully or error
func (ex *AddCommentExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute AddComment for comment {Id: %s}\n",
		ex.In_comment_info.Comment.Id)

	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	var parentCommentId *uuid.UUID = nil

	if ex.In_comment_info.Comment.ParentCommentId != "" {
		parsed := uuid.MustParse(ex.In_comment_info.Comment.ParentCommentId)
		parentCommentId = &parsed
	}


	subredditCommentModel := model.SubredditComments{
		ID:              uuid.MustParse(ex.In_comment_info.Comment.Id),
		Content:         ex.In_comment_info.Comment.Content,
		Image:           &ex.In_comment_info.Comment.Image,
		NumberOfVotes:   0,
		OwnerHandle:     ex.In_comment_info.Comment.OwnerHandle,
		ParentCommentID: parentCommentId,
		PostID:          uuid.MustParse(ex.In_comment_info.Comment.PostId),
	}

	subredditCommentInsertStmt := SubredditComments.INSERT(SubredditComments.AllColumns).MODEL(subredditCommentModel)

	userCommentModel := model.UserComments{
		ID:              uuid.MustParse(ex.In_comment_info.Comment.Id),
		Content:         ex.In_comment_info.Comment.Content,
		Image:           &ex.In_comment_info.Comment.Image,
		NumberOfVotes:   0,
		OwnerHandle:     ex.In_comment_info.Comment.OwnerHandle,
		ParentCommentID: parentCommentId,
		PostID:          uuid.MustParse(ex.In_comment_info.Comment.PostId),
	}

	userCommentInsertStmt := UserComments.INSERT(UserComments.AllColumns).MODEL(userCommentModel)

	isPreparedTx := ex.In_comment_info.UserShard != int32(rdb.shardNum) || ex.In_comment_info.SubredditShard != int32(rdb.shardNum)
	isUserShard := ex.In_comment_info.UserShard == int32(rdb.shardNum)
	isSubredditShard := ex.In_comment_info.SubredditShard == int32(rdb.shardNum)

	if !isPreparedTx {
		log.Printf("Creating a new comment {Id: %s} as part of a non-prepared transaction", ex.In_comment_info.Comment.Id)
	} else if isUserShard {
		log.Printf("Creating a new userComment {Id: %s} as part of a prepared transaction", ex.In_comment_info.Comment.Id)
	} else {
		log.Printf("Creating a new subredditComment {Id: %s} as part of a prepared transaction", ex.In_comment_info.Comment.Id)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, errors.New("couldn't start create comment transaction")
	}
	defer tx.Rollback()

	if isSubredditShard {
		// create subreddit comment
		_, err = subredditCommentInsertStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to insert SubredditComment {Id: %s}. rolling back tx\n", ex.In_comment_info.Comment.Id)
			return false, err
		}
	}

	if isUserShard {
		// create user comment
		_, err = userCommentInsertStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to insert UserComment {Id: %s}. rolling back tx\n", ex.In_comment_info.Comment.Id)
			return false, err
		}
	}

	if isPreparedTx {
		preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.In_comment_info.TwopcInfo.TransactionId, rdb.shardNum, rdb.replicaNum)
		preparedTxRaw := fmt.Sprintf("PREPARE TRANSACTION '%s'", preparedTxId)

		preparedStmt := RawStatement(preparedTxRaw)

		_, err = preparedStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to prepare create comment tx for comment {Id: %s}.\n", ex.In_comment_info.Comment.Id)
			return false, err
		}
		tx.Commit()
		log.Printf("successfully prepared tx for creating comment {Id: %s}\n", ex.In_comment_info.Comment.Id)

	} else {
		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return false, errors.New("failed to commit create comment tx")
		}
		log.Printf("successfully commited tx for creating comment {Id: %s}\n", ex.In_comment_info.Comment.Id)
	}

	return true, nil
}
