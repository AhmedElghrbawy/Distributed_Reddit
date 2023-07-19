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

func (rdb *rdbServer) GetPost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &GetPostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_post_info.MessageInfo.Id,
		result: &PostDTO{},
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
			return replyInfo.result.(*PostDTO).mapToProto(), nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}

}

func (rdb *rdbServer) CreatePost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &CreatePostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_post_info.MessageInfo.Id,
		result: false,
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
			return in_post_info.Post, nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) GetPosts(ctx context.Context, message_info *pb.MessageInfo) (*pb.PostList, error) {
	op := Op{
		Executer: &GetPostsExecuter{},
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
			postsDtos := replyInfo.result.(*[]PostDTO)
			postsProtos := make([]*pb.Post, 0)

			for _, postDto := range *postsDtos {
				postsProtos = append(postsProtos, postDto.mapToProto())
			}
			return &pb.PostList{Posts: postsProtos}, nil

		} else {
			return nil, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) PinUnpin(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &PinUnpinPostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_post_info.MessageInfo.Id,
		result: &PostDTO{},
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
			return replyInfo.result.(*PostDTO).mapToProto(), nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) UpVote(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &ChangeVoteValueForPostExecuter{In_post_info: in_post_info, ValueToAdd: 1},
		Id:       in_post_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_post_info.MessageInfo.Id,
		result: &PostDTO{},
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
			in_post_info.Post.NumberOfVotes = replyInfo.result.(int32)
			return in_post_info.Post, nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) DownVote(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &ChangeVoteValueForPostExecuter{In_post_info: in_post_info, ValueToAdd: -1},
		Id:       in_post_info.MessageInfo.Id,
	}

	replyInfo := replyInfo{
		id:     in_post_info.MessageInfo.Id,
		result: &PostDTO{},
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
			in_post_info.Post.NumberOfVotes = replyInfo.result.(int32)
			return in_post_info.Post, nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}
