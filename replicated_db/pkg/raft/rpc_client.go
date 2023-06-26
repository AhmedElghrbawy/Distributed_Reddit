package raft

import (
	"context"
	"log"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
)

/*
 example code to send a RequestVote RPC to a server.
 server is the index of the target server in rf.peers[].
 expects RPC arguments in args.
 fills in *reply with RPC reply, so caller should
 pass &reply.
 the types of the args and reply passed to Call() must be
 the same as the types of the arguments declared in the
 handler function (including whether they are pointers).

 The labrpc package simulates a lossy network, in which servers
 may be unreachable, and in which requests and replies may be lost.
 Call() sends a request and waits for a reply. If a reply arrives
 within a timeout interval, Call() returns true; otherwise
 Call() returns false. Thus Call() may not return for a while.
 A false return can be caused by a dead server, a live server that
 can't be reached, a lost request, or a lost reply.

 Call() is guaranteed to return (perhaps after a delay) *except* if the
 handler function on the server side does not return.  Thus there
 is no need to implement your own timeouts around Call().

 look at the comments in ../labrpc/labrpc.go for more details.

 if you're having trouble getting RPC to work, check that you've
 capitalized all field names in structs passed over RPC, and
 that the caller passes the address of the reply struct with &, not
 the struct itself.
*/

const DurationTO = 5 * time.Second

// ? this methods sends a request vote rpc
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

	// Debug(dLeader, "S%d sending append entries to: %d, args: %s\n", rf.me, server, args)
	return reply, nil
}
