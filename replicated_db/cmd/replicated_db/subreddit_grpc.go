package main

import (
	"bytes"
	context "context"
	"encoding/gob"
	"errors"
	"log"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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
		return replyInfo.result.(*SubredditDTO).mapToProto(), nil
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}

}

func (rdb *rdbServer) CreateSubreddit(ctx context.Context, in_subreddit *pb.SubredditInfo) (*pb.Subreddit, error) {
	return nil, nil
}

func (rdb *rdbServer) GetSubreddits(ctx context.Context, in *emptypb.Empty) (*pb.SubredditList, error) {
	return nil, nil
}

func (rdb *rdbServer) Temp() {
	time.Sleep(5 * time.Second)
	log.Println("starting to execute temp")

	in_subreddit_info := &pb.SubredditInfo{
		Subreddit: &pb.Subreddit{
			Handle: "English",
		},
	}

	op := Op{
		Executer: &GetSubredditExecuter{In_subreddit_info: in_subreddit_info},
		Id:       "yeeet",
	}

	replyInfo := replyInfo{
		id:     "yeeet",
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
		log.Println("not the leader")
	}

	select {
	case <-replyInfo.ch:
		// return replyInfo.result.(*SubredditDTO).mapToProto(), nil
		log.Printf("result returned: %v\n", replyInfo.result.(*SubredditDTO).mapToProto())
	case <-time.After(5 * time.Second): // ? magic number
		// return nil, errors.New("timed out")
		log.Printf("timed out")
	}

}
