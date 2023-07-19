package main

import (
	context "context"
	"errors"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (rdb *rdbServer) GetUser(ctx context.Context, in_user_info *pb.UserInfo) (*pb.User, error) {
	op := Op{
		Executer: &GetUserExecuter{In_user_info: in_user_info},
		Id:       in_user_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, errors.New("not the leader")
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return replyInfo.result.(*UserDTO).mapToProto(), nil
		} else {
			return &pb.User{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) CreateUser(ctx context.Context, in_user_info *pb.UserInfo) (*pb.User, error) {
	op := Op{
		Executer: &CreateUserExecuter{In_user_info: in_user_info},
		Id:       in_user_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, errors.New("not the leader")
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return in_user_info.User, nil
		} else {
			return &pb.User{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) IncreaseKarma(ctx context.Context, in_user_info *pb.UserInfo) (*pb.User, error) {
	op := Op{
		Executer: &ChangeKarmaValueForUserExecuter{In_user_info: in_user_info, ValueToAdd: 1},
		Id:       in_user_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, errors.New("not the leader")
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			in_user_info.User.Karma = replyInfo.result.(int32)
			return in_user_info.User, nil
		} else {
			return &pb.User{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) DecreaseKarma(ctx context.Context, in_user_info *pb.UserInfo) (*pb.User, error) {
	op := Op{
		Executer: &ChangeKarmaValueForUserExecuter{In_user_info: in_user_info, ValueToAdd: -1},
		Id:       in_user_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, errors.New("not the leader")
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			in_user_info.User.Karma = replyInfo.result.(int32)
			return in_user_info.User, nil
		} else {
			return &pb.User{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) Follow(ctx context.Context, user_followage_info *pb.UserFollowage) (*emptypb.Empty, error) {
	return nil, nil
}

func (rdb *rdbServer) Unfollow(ctx context.Context, user_followage_info *pb.UserFollowage) (*emptypb.Empty, error) {
	return nil, nil
}

func (rdb *rdbServer) JoinSubreddit(ctx context.Context, membership_info *pb.UserSubredditMembership) (*emptypb.Empty, error) {
	return nil, nil
}

func (rdb *rdbServer) LeaveSubreddit(ctx context.Context, membership_info *pb.UserSubredditMembership) (*emptypb.Empty, error) {
	return nil, nil
}
