package raft

import (
	"context"
	"log"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
)

const DurationTO = 2 * time.Second

func (rf *Raft) sendRequestVote(server int, args *pb.RequestVoteArgs) (*pb.RequestVoteReply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	reply, err := rf.peers[server].RequestVote(ctx, args)

	if err != nil {
		log.Printf("couldn't send request vote %v\n", err)
		return nil, err
	}

	return reply, nil
}

func (rf *Raft) sendAppendEntries(server int, args *pb.AppendEntriesArgs) (*pb.AppendEntriesReply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DurationTO)
	defer cancel()

	reply, err := rf.peers[server].AppendEntries(ctx, args)

	if err != nil {
		log.Printf("couldn't send request append entries %v\n", err)
		return nil, err
	}

	// logger.Debug(logger.DLeader, "S%d sending append entries to: %d, args: %s\n", rf.me, server, args)
	return reply, nil
}
