package main

import (
	"bytes"
	context "context"
	"encoding/gob"
	"errors"
	"log"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) GetSubreddit(ctx context.Context, in_subreddit_info *pb.SubredditInfo) (*pb.Subreddit, error) {
	op := Op{
		Executer: &GetSubredditExecuter{In_subreddit_info: in_subreddit_info},
		Id:       in_subreddit_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_subreddit_info.MessageInfo.Id,
		result: &SubredditDTO{},
		err:    &CommandNotExecutedError{},
		ch:     make(chan struct{}),
	}

	rdb.mu.Lock()
	rdb.replyMap[replyInfo.id] = &replyInfo
	rdb.mu.Unlock()

	var encodedOp bytes.Buffer
	enc := gob.NewEncoder(&encodedOp)

	err := enc.Encode(op)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	_, _, isLeader := rdb.rf.Start(encodedOp.Bytes())

	if !isLeader {
		rdb.mu.Lock()
		delete(rdb.replyMap, replyInfo.id)
		rdb.mu.Unlock()
		// TODO: return a reply status that indicates wrong leader
		return nil, errors.New("not the leader")
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

	replyInfo := replyInfo{
		id:     in_subreddit_info.MessageInfo.Id,
		result: &SubredditDTO{},
		err:    &CommandNotExecutedError{},
		ch:     make(chan struct{}),
	}

	rdb.mu.Lock()
	rdb.replyMap[replyInfo.id] = &replyInfo
	rdb.mu.Unlock()

	var encodedOp bytes.Buffer
	enc := gob.NewEncoder(&encodedOp)

	err := enc.Encode(op)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	_, _, isLeader := rdb.rf.Start(encodedOp.Bytes())

	if !isLeader {
		rdb.mu.Lock()
		delete(rdb.replyMap, replyInfo.id)
		rdb.mu.Unlock()
		// TODO: return a reply status that indicates wrong leader
		return nil, errors.New("not the leader")
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

	replyInfo := replyInfo{
		id:     message_info.Id,
		result: nil,
		err:    &CommandNotExecutedError{},
		ch:     make(chan struct{}),
	}

	rdb.mu.Lock()
	rdb.replyMap[replyInfo.id] = &replyInfo
	rdb.mu.Unlock()

	var encodedOp bytes.Buffer
	enc := gob.NewEncoder(&encodedOp)

	err := enc.Encode(op)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	_, _, isLeader := rdb.rf.Start(encodedOp.Bytes())

	if !isLeader {
		rdb.mu.Lock()
		delete(rdb.replyMap, replyInfo.id)
		rdb.mu.Unlock()
		// TODO: return a reply status that indicates wrong leader
		return nil, errors.New("not the leader")
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
