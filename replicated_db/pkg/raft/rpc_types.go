package raft

import "fmt"

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

func (args *RequestVoteArgs) String() string {
	return fmt.Sprintf("{Term: %d, candidateId: %d, LastLogIndex: %d, lastLogTerm: %d}", args.Term, args.CandidateId, args.LastLogIndex, args.LastLogTerm)
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []logEntry
	LeaderCommit int
	IsHertbeat   bool
}

func (args *AppendEntriesArgs) String() string {
	return fmt.Sprintf("{Term: %d, LeaderId: %d, prevLogIndex: %d, preLogTerm: %d, Entries: %v, leaderCommit: %d, IsHeartBeat: %t}", args.Term, args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, args.Entries, args.LeaderCommit, args.IsHertbeat)
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	XTerm   int
	XIndex  int
	XLen    int
}


type rpcType string

const (
	APPENT rpcType = "AE"
	REQVOT rpcType = "RV"
	HERTBT rpcType = "HB"
)