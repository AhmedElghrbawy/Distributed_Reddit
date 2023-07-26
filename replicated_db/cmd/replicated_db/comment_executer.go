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

var grpc_column_to_subcomment_column = map[pb.CommentUpdatedColumn]Column{
	pb.CommentUpdatedColumn_COMMENT_CONTENT:         SubredditComments.Content,
	pb.CommentUpdatedColumn_COMMENT_IMAGE:           SubredditComments.Image,
	pb.CommentUpdatedColumn_COMMENT_NUMBER_OF_VOTES: SubredditComments.NumberOfVotes,
}

var grpc_column_to_usercomment_column = map[pb.CommentUpdatedColumn]Column{
	pb.CommentUpdatedColumn_COMMENT_CONTENT:         AliasedUserComments.Content,
	pb.CommentUpdatedColumn_COMMENT_IMAGE:           AliasedUserComments.Image,
	pb.CommentUpdatedColumn_COMMENT_NUMBER_OF_VOTES: AliasedUserComments.NumberOfVotes,
}

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

// returns (*CommentDTO, error): the comment after update or error
type UpdateCommentExecuter struct {
	In_comment_info      *pb.CommentInfo
	Updated_column_value map[pb.CommentUpdatedColumn]interface{}
}

func (ex *UpdateCommentExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	ex.Updated_column_value = map[pb.CommentUpdatedColumn]interface{}{
		pb.CommentUpdatedColumn_COMMENT_CONTENT:         ex.In_comment_info.Comment.Content,
		pb.CommentUpdatedColumn_COMMENT_IMAGE:           ex.In_comment_info.Comment.Image,
		pb.CommentUpdatedColumn_COMMENT_NUMBER_OF_VOTES: Int(int64(ex.In_comment_info.Comment.NumberOfVotes)),
	}

	subcomment_columns_to_update := ColumnList{}
	for _, c := range ex.In_comment_info.UpdatedColumns {
		subcomment_columns_to_update = append(subcomment_columns_to_update, grpc_column_to_subcomment_column[c])
	}

	usercomment_columns_to_update := ColumnList{}
	for _, c := range ex.In_comment_info.UpdatedColumns {
		usercomment_columns_to_update = append(usercomment_columns_to_update, grpc_column_to_usercomment_column[c])
	}

	subcomment_updated_values := make([]interface{}, 0)

	for _, c := range ex.In_comment_info.UpdatedColumns {
		if c == pb.CommentUpdatedColumn_COMMENT_NUMBER_OF_VOTES {
			value := ex.Updated_column_value[c].(IntegerExpression).ADD(SubredditComments.NumberOfVotes)
			subcomment_updated_values = append(subcomment_updated_values, value)
		} else {
			subcomment_updated_values = append(subcomment_updated_values, ex.Updated_column_value[c])
		}
	}

	usercomment_updated_values := make([]interface{}, 0)

	for _, c := range ex.In_comment_info.UpdatedColumns {
		if c == pb.CommentUpdatedColumn_COMMENT_NUMBER_OF_VOTES {
			value := ex.Updated_column_value[c].(IntegerExpression).ADD(AliasedUserComments.NumberOfVotes)
			usercomment_updated_values = append(usercomment_updated_values, value)
		} else {
			usercomment_updated_values = append(usercomment_updated_values, ex.Updated_column_value[c])
		}
	}
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to update post {Id: %s} columns: %v, values: %v\n", ex.In_comment_info.Comment.Id, usercomment_columns_to_update, usercomment_updated_values)

	isUserComment := ex.In_comment_info.UserShard == int32(rdb.shardNum)
	isSubredditComment := ex.In_comment_info.SubredditShard == int32(rdb.shardNum)

	subCommentStmt := SubredditComments.
		UPDATE(subcomment_columns_to_update).
		SET(subcomment_updated_values[0], subcomment_updated_values[1:]...).
		WHERE(CAST(table.SubredditComments.ID).AS_TEXT().EQ(String(ex.In_comment_info.Comment.Id))).
		RETURNING(SubredditComments.AllColumns)

	userCommentStmt := AliasedUserComments.
		UPDATE(usercomment_columns_to_update).
		SET(usercomment_updated_values[0], usercomment_updated_values[1:]...).
		WHERE(CAST(AliasedUserComments.ID).AS_TEXT().EQ(String(ex.In_comment_info.Comment.Id))).
		RETURNING(AliasedUserComments.AllColumns)

	result := CommentDTO{}

	if isUserComment {
		err = userCommentStmt.Query(db, &result)

		if err != nil {
			log.Printf("update comment {Id: %s} command failed %v\n", ex.In_comment_info.Comment.Id, err)
			return 0, err
		}
	}

	if isSubredditComment {
		err = subCommentStmt.Query(db, &result)

		if err != nil {
			log.Printf("update {Id: %s} command failed %v\n", ex.In_comment_info.Comment.Id, err)
			return 0, err
		}
	}

	log.Printf("change vote value for command completed\n")

	return &result, nil

}
