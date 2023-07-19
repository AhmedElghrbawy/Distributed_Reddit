package main

import (
	"database/sql"
	"log"

	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	. "github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	. "github.com/go-jet/jet/v2/postgres"
)

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

// returns (int, error): the new karma value for the user or error
type ChangeKarmaValueForUserExecuter struct {
	In_user_info *pb.UserInfo
	ValueToAdd   int
}

func (ex *ChangeKarmaValueForUserExecuter) Execute(rdb *rdbServer) (interface{}, error) {
	db, err := sql.Open("postgres", rdb.dbConnectionStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Printf("Preparing to execute change karma for User {Id: %s}\n",
		ex.In_user_info.User.Handle)

	curKarmaStmt := SELECT(Users.Karma).FROM(Users).WHERE(Users.Handle.EQ(String(ex.In_user_info.User.Handle)))

	curUser := UserDTO{}
	err = curKarmaStmt.Query(db, &curUser)

	if err != nil {
		log.Printf("change karma value for User {handle: %s} command failed %v\n", ex.In_user_info.User.Handle, err)
		return 0, err
	}
	curKarma := curUser.Karma

	u := model.Users{
		Karma: curKarma + int32(ex.ValueToAdd),
	}

	updateKarmaStmt := Users.
		UPDATE(Users.Karma).
		MODEL(u).
		WHERE(Users.Handle.EQ(String(ex.In_user_info.User.Handle))).
		RETURNING(Users.Karma)

	result := UserDTO{}

	err = updateKarmaStmt.Query(db, &result)

	if err != nil {
		log.Printf("change karma value for User {handle: %s} command failed %v\n", ex.In_user_info.User.Handle, err)
		return 0, err
	}

	log.Printf("change karma value for user command completed\n")
	return result.Karma, nil
}
