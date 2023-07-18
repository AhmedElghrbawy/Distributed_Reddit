package main

import (
	"bytes"
	context "context"
	"encoding/gob"
	"errors"
	"log"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (rdb *rdbServer) Commit(ctx context.Context, twopc_info *pb.TwoPhaseCommitInfo) (*emptypb.Empty, error) {
	op := Op{
		Executer: &CommitExecuter{Twopc_info: twopc_info},
		Id:       twopc_info.TransactionId,
	}

	replyInfo := replyInfo{
		id:     twopc_info.TransactionId,
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
			return &emptypb.Empty{}, nil
		} else {
			return &emptypb.Empty{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) Rollback(ctx context.Context, twopc_info *pb.TwoPhaseCommitInfo) (*emptypb.Empty, error) {
	op := Op{
		Executer: &RollbackExecuter{Twopc_info: twopc_info},
		Id:       twopc_info.TransactionId,
	}

	replyInfo := replyInfo{
		id:     twopc_info.TransactionId,
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
			return &emptypb.Empty{}, nil
		} else {
			return &emptypb.Empty{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}
