package raft

import (
	"context"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/logger"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
)

func (rf *Raft) RequestVote(ctx context.Context, args *pb.RequestVoteArgs) (*pb.RequestVoteReply, error) {
	reply := &pb.RequestVoteReply{}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	logger.Debug(logger.DTrace, "S%d got a request vote rpc from S%d. currentTerm: %d, args: %s\n", rf.me, args.CandidateId, rf.currentTerm, args)

	if rf.currentTerm > int(args.Term) {
		logger.Debug(logger.DDrop, "S%d rejecting a request vote with outdated term. rpc term: %d, currentTerm: %d\n", rf.me, args.Term, rf.currentTerm)
		reply.Term = int32(rf.currentTerm)
		reply.VoteGranted = false
		return reply, nil
	}

	if rf.currentTerm < int(args.Term) {
		logger.Debug(logger.DElection, "S%d updataing term from %d to %d after getting RV from %d.\n", rf.me, rf.currentTerm, args.Term, args.CandidateId)
		revertToFollower(rf, int(args.Term), -1)
	}

	reply.Term = int32(rf.currentTerm)
	reply.VoteGranted = false

	if rf.votedFor == -1 && rf.leaderId == -1 {
		n := len(rf.log)
		// election restriction
		if args.LastLogTerm < int32(rf.log[n-1].Term) {
			logger.Debug(logger.DDrop, "S%d rejecting RV. Election restrection. my lastLogTerm is larger than args.LastLogTerm me: %s, args: %s\n", rf.me, rf, args)
			return reply, nil
		}
		if int(args.LastLogTerm) > rf.log[n-1].Term || int(args.LastLogIndex) >= n-1 {
			// logger.Debug(logger.DInfo, "S%d received a valid rpc. me: %s, args: %s\n", rf.me, rf, args)
			rf.lastRpcTime = time.Now()

			logger.Debug(logger.DVote, "S%d is granting vote for %d. Term: %d, state: %s\n", rf.me, args.CandidateId, rf.currentTerm, rf.state)
			rf.votedFor = int(args.CandidateId)
			reply.VoteGranted = true
			// TODO: you might need to optimize here. rf.persist() is called twice, once in revertToFollower and here which might take a lot of time
			rf.persist()

		} else {
			logger.Debug(logger.DDrop, "S%d rejecting RV. Election restrection. my log length is larger. me: %s, args: %s\n", rf.me, rf, args)
		}
	} else {
		logger.Debug(logger.DElection, "S%d already voted for S%d. droping RV from S%d in term: %d. or there exists a leader. leaderId: %d\n", rf.me, rf.votedFor, args.CandidateId, rf.currentTerm, rf.leaderId)
	}

	return reply, nil
}

func (rf *Raft) AppendEntries(ctx context.Context, args *pb.AppendEntriesArgs) (*pb.AppendEntriesReply, error) {
	reply := &pb.AppendEntriesReply{}
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.currentTerm > int(args.Term) {
		reply.Term = int32(rf.currentTerm)
		reply.Success = false
		return reply, nil
	}

	if rf.currentTerm == int(args.Term) && rf.state == leader && rf.leaderId == rf.me {
		logger.Debug(logger.DError, "S%d Two leaders detected for the same term. Term = %d, peers = (%d, %d)\n", rf.me, rf.currentTerm, rf.me, args.LeaderId)
	}

	rf.lastRpcTime = time.Now()

	logger.Debug(logger.DInfo, "S%d got AE from %d. me: %s, args: %s\n", rf.me, args.LeaderId, rf, args)

	if rf.currentTerm <= int(args.Term) {
		if rf.currentTerm < int(args.Term) {
			logger.Debug(logger.DElection, "S%d updataing term from %d to %d after getting AE from %d.\n", rf.me, rf.currentTerm, args.Term, args.LeaderId)
		}
		revertToFollower(rf, int(args.Term), int(args.LeaderId))
	}

	reply.Term = int32(rf.currentTerm)

	if int(args.PrevLogIndex) >= len(rf.log) || rf.log[args.PrevLogIndex].Term != int(args.PrevLogTerm) {
		logger.Debug(logger.DDrop, "S%d droping AE. Consistency check failed. me: %s, args: %s\n", rf.me, rf, args)
		reply.Success = false
		reply.XLen = int32(len(rf.log))

		if int(args.PrevLogIndex) >= len(rf.log) {
			reply.XTerm = -1
			reply.XIndex = -1
			return reply, nil
		}

		reply.XTerm = int32(rf.log[args.PrevLogIndex].Term)

		for i := 0; i < len(rf.log); i++ {
			if rf.log[i].Term == int(reply.XTerm) {
				reply.XIndex = int32(i)
				return reply, nil
			}
		}
		// logger.Debug(logger.DError, "S%d Xindex couldn't be found.\n", rf.me)
		return reply, nil
	}

	// Figure 2: A. AppendEntries RPC: rec implementaion: 3. If an existing entry conflicts with a new one (same index but different term) delete

	logger.Debug(logger.DLog, "S%d about to accept logs. me: %s, args: %s\n", rf.me, rf, args)
	// TODO: you might need to optimize here. rf.persist() is called twice, once in revertToFollower and here which might take a lot of time
	for i, entry := range args.Entries {
		if int(entry.Index) < len(rf.log) {
			if rf.log[entry.Index].Term != int(entry.Term) {
				rf.log = append(rf.log[:entry.Index], MapGrpcEntriesToRaftEntries(args.Entries)[i:]...)
				rf.persist()
				break
			}
		} else {
			rf.log = append(rf.log, MapGrpcEntriesToRaftEntries(args.Entries)[i:]...)
			rf.persist()
			break
		}
	}
	logger.Debug(logger.DLog, "S%d accepted logs. me: %s, args: %s\n", rf.me, rf, args)

	reply.Success = true

	if int(args.LeaderCommit) > rf.commitIndex {
		if len(rf.log)-1 < int(args.LeaderCommit) {
			rf.commitIndex = len(rf.log) - 1
		} else {
			rf.commitIndex = int(args.LeaderCommit)
		}
		go applySafeEntries(rf)
	}
	return reply, nil
}
