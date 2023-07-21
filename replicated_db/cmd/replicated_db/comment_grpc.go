package main

import (
	context "context"
	"errors"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) AddComment(ctx context.Context, in_comment_info *pb.CommentInfo) (*pb.Comment, error) {
	op := Op{
		Executer: &AddCommentExecuter{In_comment_info: in_comment_info},
		Id:       in_comment_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, errors.New("not the leader")
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return in_comment_info.Comment, nil
		} else {
			return &pb.Comment{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) UpVoteComment(ctx context.Context, in_comment_info *pb.CommentInfo) (*pb.Comment, error) {
	return nil, nil
}

func (rdb *rdbServer) DownVoteComment(ctx context.Context, in_comment_info *pb.CommentInfo) (*pb.Comment, error) {
	return nil, nil
}
