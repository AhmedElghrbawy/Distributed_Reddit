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
)

var grpc_column_to_user_column = map[pb.UserUpdatedColumn]Column{
	pb.UserUpdatedColumn_AVATAR:       Users.Avatar,
	pb.UserUpdatedColumn_DISPLAY_NAME: Users.DisplayName,
	pb.UserUpdatedColumn_KARMA:        Users.Karma,
}

// returns (*UserDto, error): the user requested or error
type GetUserExecuter struct {
	In_user_info *pb.UserInfo
}

func (ex *GetUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute GetUser for User {Id: %s}\n",
		ex.In_user_info.User.Handle)

	stmt := SELECT(
		Users.AllColumns, AliasedUserPosts.AllColumns, AliasedUserComments.AllColumns,
		PostTags.TagName, SubredditUsers.AllColumns, Followage.AllColumns,
	).FROM(Users.
		LEFT_JOIN(AliasedUserPosts, AliasedUserPosts.OwnerHandle.EQ(Users.Handle)).
		LEFT_JOIN(AliasedUserComments, AliasedUserComments.PostID.EQ(AliasedUserPosts.ID)).
		LEFT_JOIN(PostTags, PostTags.PostID.EQ(AliasedUserPosts.ID)).
		LEFT_JOIN(SubredditUsers, SubredditUsers.UserHandle.EQ(Users.Handle)).
		LEFT_JOIN(Followage, Followage.FollowedHandle.EQ(Users.Handle).OR(Followage.FollowerHandle.EQ(Users.Handle))),
	).WHERE(
		Users.Handle.EQ(String(ex.In_user_info.User.Handle)),
	)

	result := UserDTO{}

	err = stmt.Query(db, &result)

	if err != nil {
		log.Printf("GetUser {Handle: %s} command failed %v\n", ex.In_user_info.User.Handle, err)
		return &SubredditDTO{}, err
	}

	log.Printf("GetUser {Handle: %s} command completed\n", ex.In_user_info.User.Handle)

	return &result, nil
}

// returns (bool, error): true if user created successfully or error
type CreateUserExecuter struct {
	In_user_info *pb.UserInfo
}

func (ex *CreateUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute CreateUser for User {Id: %s}\n",
		ex.In_user_info.User.Handle)

	userModel := model.Users{
		Handle:      ex.In_user_info.User.Handle,
		DisplayName: ex.In_user_info.User.DisplayName,
		Avatar:      ex.In_user_info.User.Avatar,
		Karma:       ex.In_user_info.User.Karma,
		CreatedAt:   ex.In_user_info.User.CreatedAt.AsTime(),
	}

	userInsertStmt := Users.INSERT(Users.AllColumns).
		MODEL(userModel)

	_, err = userInsertStmt.Exec(db)

	if err != nil {
		log.Printf("failed to insert user {handle: %s}.\n", ex.In_user_info.User.Handle)
		return false, err
	}

	log.Printf("create user command completed\n")
	return true, nil
}

// returns (bool, error): true if operation executed successfully or error
type FollowUnfollowUserExecuter struct {
	User_followage_info *pb.UserFollowage
	Follow              bool
}

func (ex *FollowUnfollowUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	folloawgeModel := model.Followage{
		FollowerHandle: ex.User_followage_info.FromHandle,
		FollowedHandle: ex.User_followage_info.ToHandle,
	}
	insertStmt := Followage.INSERT(Followage.FollowedHandle, Followage.FollowerHandle).
		MODEL(folloawgeModel).
		ON_CONFLICT(Followage.FollowedHandle, Followage.FollowerHandle).DO_NOTHING()
	deleteStmt := Followage.DELETE().
		WHERE(Followage.FollowerHandle.EQ(String(ex.User_followage_info.FromHandle)).
			AND(Followage.FollowedHandle.EQ(String(ex.User_followage_info.ToHandle))))

	isPreparedTx := ex.User_followage_info.FromShard != ex.User_followage_info.ToShard

	if isPreparedTx {
		log.Printf("Preparing to execute follow/unfollow {From: %s, to: %s} as part of prepared tx\n",
			ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return false, errors.New("couldn't start transaction")
		}
		defer tx.Rollback()

		if ex.Follow {
			_, err = insertStmt.ExecContext(ctx, tx)

		} else {
			_, err = deleteStmt.ExecContext(ctx, tx)
		}

		if err != nil {
			log.Printf("failed to follow/unfollow {From: %s, to: %s}. rolling back tx\n", ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)
			return false, err
		}

		preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.User_followage_info.TwopcInfo.TransactionId, rdb.shardNum, rdb.replicaNum)
		preparedTxRaw := fmt.Sprintf("PREPARE TRANSACTION '%s'", preparedTxId)

		preparedStmt := RawStatement(preparedTxRaw)

		_, err = preparedStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to prepare follow/unfollow {From: %s, to: %s} tx.\n", ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)
			return false, err
		}

		tx.Commit()
		log.Printf("successfully prepared tx for follow/unfollow {From: %s, to: %s} tx\n", ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)
	} else {
		log.Printf("Preparing to execute follow/unfollow {From: %s, to: %s} as part of non-prepared\n",
			ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)

		if ex.Follow {
			_, err = insertStmt.Exec(db)

		} else {
			_, err = deleteStmt.Exec(db)
		}
		if err != nil {
			log.Printf("failed to follow/unfollow {From: %s, to: %s}\n", ex.User_followage_info.FromHandle, ex.User_followage_info.ToHandle)
			return false, err
		}
	}

	log.Printf("follow/unfollow command completed\n")
	return true, nil
}

// returns (bool, error): true if operation executed successfully or error
type JoinLeaveSubredditUserExecuter struct {
	UserSubredditMembership *pb.UserSubredditMembership
	Join                    bool
}

func (ex *JoinLeaveSubredditUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	subredditUserModel := model.SubredditUsers{
		UserHandle:      ex.UserSubredditMembership.UserHandle,
		SubredditHandle: ex.UserSubredditMembership.SubredditHandle,
		IsAdmin:         false,
	}
	insertStmt := SubredditUsers.INSERT(SubredditUsers.AllColumns).
		MODEL(subredditUserModel).
		ON_CONFLICT(SubredditUsers.UserHandle, SubredditUsers.SubredditHandle).DO_NOTHING()
	deleteStmt := SubredditUsers.DELETE().
		WHERE(SubredditUsers.UserHandle.EQ(String(ex.UserSubredditMembership.UserHandle)).
			AND(SubredditUsers.SubredditHandle.EQ(String(ex.UserSubredditMembership.SubredditHandle))))

	isPreparedTx := ex.UserSubredditMembership.UserShard != ex.UserSubredditMembership.SubredditShard

	if isPreparedTx {
		log.Printf("Preparing to execute join/leave {User: %s, Subreddit: %s} as part of prepared tx\n",
			ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return false, errors.New("couldn't start transaction")
		}
		defer tx.Rollback()

		if ex.Join {
			_, err = insertStmt.ExecContext(ctx, tx)

		} else {
			_, err = deleteStmt.ExecContext(ctx, tx)
		}

		if err != nil {
			log.Printf("failed to join/leave {User: %s, Subreddit: %s}. rolling back tx\n", ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)
			return false, err
		}

		preparedTxId := fmt.Sprintf("%s-S%d_R%d", ex.UserSubredditMembership.TwopcInfo.TransactionId, rdb.shardNum, rdb.replicaNum)
		preparedTxRaw := fmt.Sprintf("PREPARE TRANSACTION '%s'", preparedTxId)

		preparedStmt := RawStatement(preparedTxRaw)

		_, err = preparedStmt.ExecContext(ctx, tx)

		if err != nil {
			log.Printf("failed to prepare join/leave {User: %s, Subreddit: %s} tx.\n", ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)
			return false, err
		}

		tx.Commit()
		log.Printf("successfully prepared tx for join/leave {User: %s, Subreddit: %s} tx\n", ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)
	} else {
		log.Printf("Preparing to execute join/leave {User: %s, Subreddit: %s} as part of non-prepared\n",
			ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)

		if ex.Join {
			_, err = insertStmt.Exec(db)

		} else {
			_, err = deleteStmt.Exec(db)
		}
		if err != nil {
			log.Printf("failed to join/leave {User: %s, Subreddit: %s}\n", ex.UserSubredditMembership.UserHandle, ex.UserSubredditMembership.SubredditHandle)
			return false, err
		}
	}

	log.Printf("join/leave subreddit command completed\n")
	return true, nil
}

// returns (*UserDTO, error): the user after updating or error
type UpdateUserExecuter struct {
	In_user_info         *pb.UserInfo
	Updated_column_value map[pb.UserUpdatedColumn]interface{}
}

func (ex *UpdateUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	ex.Updated_column_value = map[pb.UserUpdatedColumn]interface{}{
		pb.UserUpdatedColumn_AVATAR:       ex.In_user_info.User.Avatar,
		pb.UserUpdatedColumn_DISPLAY_NAME: ex.In_user_info.User.DisplayName,
		pb.UserUpdatedColumn_KARMA:        Int(int64(ex.In_user_info.User.Karma)),
	}

	user_columns_to_update := ColumnList{}
	for _, c := range ex.In_user_info.UpdatedColumns {
		user_columns_to_update = append(user_columns_to_update, grpc_column_to_user_column[c])
	}

	user_updated_values := make([]interface{}, 0)

	for _, c := range ex.In_user_info.UpdatedColumns {
		if c == pb.UserUpdatedColumn_KARMA {
			value := ex.Updated_column_value[c].(IntegerExpression).ADD(Users.Karma)
			user_updated_values = append(user_updated_values, value)
		} else {
			user_updated_values = append(user_updated_values, ex.Updated_column_value[c])
		}
	}

	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to update user {handle: %s} columns: %v, values: %v\n", ex.In_user_info.User.Handle, user_columns_to_update, user_updated_values)

	updateStmt := Users.
		UPDATE(user_columns_to_update).
		SET(user_updated_values[0], user_updated_values[1:]...).
		WHERE(Users.Handle.EQ(String(ex.In_user_info.User.Handle))).
		RETURNING(Users.AllColumns)

	result := UserDTO{}

	err = updateStmt.Query(db, &result)

	if err != nil {
		log.Printf("update User {handle: %s} command failed %v\n", ex.In_user_info.User.Handle, err)
		return 0, err
	}

	log.Printf("update user command completed\n")
	return &result, nil
}
