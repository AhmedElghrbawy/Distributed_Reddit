package raft

import "time"

// example RequestVote RPC handler.
// ? this method handles/receives request vote rpc
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()
	// Debug(dTrace, "S%d got a request vote rpc from S%d. currentTerm: %d, args: %s\n", rf.me, args.CandidateId, rf.currentTerm, args)
	if args.CandidateId == rf.me {
		// Debug(dWarn, "S%d got request vote from myself\n", rf.me)
	}

	if rf.currentTerm > args.Term {
		// Debug(dDrop, "S%d rejecting a request vote with outdated term. rpc term: %d, currentTerm: %d\n", rf.me, args.Term, rf.currentTerm)
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	if rf.currentTerm < args.Term {
		// Debug(dElection, "S%d updataing term from %d to %d after getting RV from %d.\n", rf.me, rf.currentTerm, args.Term, args.CandidateId)
		revertToFollower(rf, args.Term, -1)
	}

	if rf.votedFor == -1 && rf.leaderId == -1 {
		n := len(rf.log)
		// election restriction
		if args.LastLogTerm < rf.log[n-1].Term {
			// Debug(dDrop, "S%d rejecting RV. Election restrection. my lastLogTerm is larger than args.LastLogTerm me: %s, args: %s\n", rf.me, rf, args)
			return
		}
		if args.LastLogTerm > rf.log[n-1].Term || args.LastLogIndex >= n-1 {
			// Debug(dInfo, "S%d received a valid rpc. me: %s, args: %s\n", rf.me, rf, args)
			rf.lastRpcTime = time.Now()

			// Debug(dVote, "S%d is granting vote for %d. Term: %d, state: %s\n", rf.me, args.CandidateId, rf.currentTerm, rf.state)
			rf.votedFor = args.CandidateId
			reply.Term = rf.currentTerm
			reply.VoteGranted = true
			// TODO: you might need to optimize here. rf.persist() is called twice, once in revert to follower and here which might take a lot of time
			rf.persist()

		} else {
			// Debug(dDrop, "S%d rejecting RV. Election restrection. my log length is larger. me: %s, args: %s\n", rf.me, rf, args)
		}
		return
	} else {
		// Debug(dElection, "S%d already voted for S%d. droping RV from S%d in term: %d. or there exists a leader. leaderId: %d\n", rf.me, rf.votedFor, args.CandidateId, rf.currentTerm, rf.leaderId)
	}

}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.LeaderId == rf.me {
		// Debug(dWarn, "S%d got append entries from myself\n")
	}
	if rf.currentTerm > args.Term {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	if rf.currentTerm == args.Term && rf.state == leader && rf.leaderId == rf.me {
		// DPrintf("*******Two leaders detected for the same term. term = %d peers = (%d, %d)*******\n", rf.currentTerm, rf.me, args.LeaderId)
		// Debug(dError, "S%d Two leaders detected for the same term. Term = %d, peers = (%d, %d)\n", rf.me, rf.currentTerm, rf.me, args.LeaderId)
	}

	rf.lastRpcTime = time.Now()

	// Debug(dInfo, "S%d got AE from %d. me: %s, args: %s\n", rf.me, args.LeaderId, rf, args)

	if rf.currentTerm <= args.Term {
		if rf.currentTerm < args.Term {
			// Debug(dElection, "S%d updataing term from %d to %d after getting AE from %d.\n", rf.me, rf.currentTerm, args.Term, args.LeaderId)
		}
		revertToFollower(rf, args.Term, args.LeaderId)
	}

	reply.Term = rf.currentTerm

	// if args.LeaderCommit < rf.commitIndex {
	// 	// Debug(dInfo, "S%d leader commit is smaller than my commit index.\n", rf.me, rf, args)
	// 	// // Debug(dInfo, "S%d Ignoring request....\n", rf.me)

	// 	// return
	// }
	if args.PrevLogIndex >= len(rf.log) || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		// Debug(dDrop, "S%d droping AE. Consistency check failed. me: %s, args: %s\n", rf.me, rf, args)
		reply.Success = false
		reply.XLen = len(rf.log)

		if args.PrevLogIndex >= len(rf.log) {
			reply.XTerm = -1
			reply.XIndex = -1
			return
		}

		reply.XTerm = rf.log[args.PrevLogIndex].Term

		for i := 0; i < len(rf.log); i++ {
			if rf.log[i].Term == reply.XTerm {
				reply.XIndex = i
				return
			}
		}
		// Debug(dError, "S%d Xindex couldn't be found.\n", rf.me)
		return
	}

	// Figure 2: A. AppendEntries RPC: rec implementaion: 3. If an existing entry conflicts with a new one (same index but different term) delete

	// Debug(dLog, "S%d about to accept logs. me: %s, args: %s\n", rf.me, rf, args)
	// TODO: you might need to optimize here. rf.persist() is called twice, once in revert to follower and here which might take a lot of time
	for i, entry := range args.Entries {
		if entry.Index < len(rf.log) {
			if rf.log[entry.Index].Term != entry.Term {
				rf.log = append(rf.log[:entry.Index], args.Entries[i:]...)
				rf.persist()
				break
			}
		} else {
			rf.log = append(rf.log, args.Entries[i:]...)
			rf.persist()
			break
		}
	}
	// rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...)
	// Debug(dLog, "S%d accepted logs. me: %s, args: %s\n", rf.me, rf, args)

	reply.Success = true

	if args.LeaderCommit > rf.commitIndex {
		if len(rf.log)-1 < args.LeaderCommit {
			rf.commitIndex = len(rf.log) - 1
		} else {
			rf.commitIndex = args.LeaderCommit
		}
		go ApplySafeEntries(rf)
	}

}
