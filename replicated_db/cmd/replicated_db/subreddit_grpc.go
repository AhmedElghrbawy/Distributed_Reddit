package main

import (
	context "context"
	"errors"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) GetSubreddit(ctx context.Context, in_subreddit_info *pb.SubredditInfo) (*pb.Subreddit, error) {
	op := Op{
		Executer: &GetSubredditExecuter{In_subreddit_info: in_subreddit_info},
		Id:       in_subreddit_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return replyInfo.result.(*SubredditDTO).mapToProto(), nil
		} else {
			return &pb.Subreddit{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}

}

func (rdb *rdbServer) CreateSubreddit(ctx context.Context, in_subreddit_info *pb.SubredditInfo) (*pb.Subreddit, error) {
	if len(in_subreddit_info.Subreddit.AdminsHandles) == 0 {
		return &pb.Subreddit{}, errors.New("subreddit must have at least one admin to be created")
	}

	op := Op{
		Executer: &CreateSubredditExecuter{In_subreddit_info: in_subreddit_info},
		Id:       in_subreddit_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return in_subreddit_info.Subreddit, nil
		} else {
			return &pb.Subreddit{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) GetSubredditsHandles(ctx context.Context, message_info *pb.MessageInfo) (*pb.SubredditList, error) {
	op := Op{
		Executer: &GetSubredditsHandlesExecuter{},
		Id:       message_info.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			subredditsDtos := replyInfo.result.(*[]SubredditDTO)
			subredditsProtos := make([]*pb.Subreddit, 0)

			for _, subredditDto := range *subredditsDtos {
				subredditsProtos = append(subredditsProtos, subredditDto.mapToProto())
			}

			return &pb.SubredditList{Subreddits: subredditsProtos}, nil
		} else {
			return nil, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}
