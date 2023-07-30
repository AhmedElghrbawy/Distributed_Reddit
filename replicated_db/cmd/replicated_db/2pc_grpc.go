package main

import (
	context "context"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (rdb *rdbServer) Commit(ctx context.Context, twopc_info *pb.TwoPhaseCommitInfo) (*emptypb.Empty, error) {
	op := Op{
		Executer: &CommitExecuter{Twopc_info: twopc_info},
		Id:       twopc_info.TransactionId,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return &emptypb.Empty{}, nil
		} else {
			return &emptypb.Empty{}, replyInfo.err
		}
	case <-time.After(time.Second):
		return nil, rdb_grpc_error_map[SERVER_RESPONSE_TIMEOUT]
	}
}

func (rdb *rdbServer) Rollback(ctx context.Context, twopc_info *pb.TwoPhaseCommitInfo) (*emptypb.Empty, error) {
	op := Op{
		Executer: &RollbackExecuter{Twopc_info: twopc_info},
		Id:       twopc_info.TransactionId,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return &emptypb.Empty{}, nil
		} else {
			return &emptypb.Empty{}, replyInfo.err
		}
	case <-time.After(time.Second):
		return nil, rdb_grpc_error_map[SERVER_RESPONSE_TIMEOUT]
	}
}
