package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
//   create a new Raft server.
// rf = Make(...)
//   start agreement on a new log entry
// rf.Start(command interface{}) (index, term, isleader)
//   ask a Raft for its current term, and whether it thinks it is leader
// rf.GetState() (term, isLeader)
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"fmt"
	"log"
	"math/rand"
	"net/netip"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/persister"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// import "bytes"
// import "../labgob"

/*
as each Raft peer becomes aware that successive log entries are
committed, the peer should send an ApplyMsg to the service (or
tester) on the same server, via the applyCh passed to Make(). set
CommandValid to true to indicate that the ApplyMsg contains a newly
committed log entry.

in Lab 3 you'll want to send other kinds of messages (e.g.,
snapshots) on the applyCh; at that point you can add fields to
ApplyMsg, but set CommandValid to false for these other uses.
*/
type ApplyMsg struct {
	CommandValid bool
	Command      []byte
	CommandIndex int
}

type logEntry struct {
	Command []byte
	Term    int
	Index   int
}

type state string

const (
	follower  state = "follower"
	candidate state = "candidate"
	leader    state = "leader"
)
const (
	MINTO  int = 600
	MAXTO  int = 1500
	HBRATE int = 100
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex           // Lock to protect shared access to this peer's state
	peers     []pb.RaftGRPCClient  // RPC end points of all peers
	persister *persister.Persister // Object to hold this peer's persisted state
	me        int                  // this peer's index into peers[]
	dead      int32                // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	pb.UnimplementedRaftGRPCServer

	// needs to be locked
	state           state
	currentTerm     int
	votedFor        int
	votesReceived   int
	leaderId        int
	electionTimeout int
	log             []logEntry
	commitIndex     int
	lastApplied     int
	nextIndex       []int
	matchIndex      []int

	applyCh         chan ApplyMsg
	appendNotifiers []chan struct{}
	lastRpcTime     time.Time
}

func (rf *Raft) String() string {
	return fmt.Sprintf("{state: %s, term: %d, log: %v, commitIndex: %d, lastApplied: %d, nextIndex: %v, matchIndex: %v, votedFor: %d, votesReceived: %d, leaderId: %d}         ", rf.state, rf.currentTerm, rf.log, rf.commitIndex, rf.lastApplied, rf.nextIndex, rf.matchIndex, rf.votedFor, rf.votesReceived, rf.leaderId)
}

/*
the service or tester wants to create a Raft server.
the ports of all the Raft servers (including this one) are in peers[].
this server's port is peers[me].
all the servers' peers[] arrays have the same order.
persister is a place for this server to save its persistent state, and also initially holds the most recent saved state, if any.
applyCh is a channel on which the tester or service expects Raft to send ApplyMsg messages.
Make() must return quickly, so it should start goroutines for any long-running work.
*/
func Make(peerAdresses []netip.AddrPort, me int,
	persister *persister.Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.persister = persister
	rf.me = me
	rf.applyCh = applyCh
	rf.appendNotifiers = make([]chan struct{}, 0)
	rf.currentTerm = -1
	rf.votedFor = -1
	rf.lastRpcTime = time.Now()
	rf.log = []logEntry{logEntry{
		Command: make([]byte, 0),
		Index:   0,
		Term:    -1,
	}}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))

	for i, peerAddress := range peerAdresses {
		if i == me {
			continue
		}
		// TODO: need to clean up these connections somewhere, maybe send a cancellation signal from the application layer
		conn, err := grpc.Dial(peerAddress.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("unable to initialize a client : %v", err)
		}
		// defer conn.Close()

		client := pb.NewRaftGRPCClient(conn)

		rf.peers[i] = client
	}

	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = 1
	}

	rf.matchIndex = make([]int, len(rf.peers))
	rf.electionTimeout = rand.Intn(MAXTO-MINTO) + MINTO

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// each peer will have a long running notifier go routine that waits to be notified of a new command or a won ellection and starts sending append entries to this peer
	// this helps avoiding inconsistencies
	for i := 0; i < len(rf.peers); i++ {
		rf.appendNotifiers = append(rf.appendNotifiers, make(chan struct{}))
		if i == rf.me {
			continue
		}
		go appendEntriesController(rf, i, rf.appendNotifiers[i])
	}

	// DPrintf("Starting peer %d\n", rf.me)
	// Debug(dInfo, "S%d Starting to operate...\n", rf.me)
	// Debug(dTimer, "S%d election time out: %dms\n", rf.me, rf.electionTimeout)
	// Your initialization code here (2A, 2B, 2C).

	startNewTerm(true, rf)
	go timer(rf)

	return rf
}

// warning: "term changed even though there were no failures". this warnning showed when running 2a tests

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	// 2A
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Debug(dInfo, "S%d Getting state... current term = %d, leader = %d state = %s\n", rf.me, rf.currentTerm, rf.leaderId, rf.state)
	return rf.currentTerm, rf.leaderId == rf.me

}

/*
the service using Raft (e.g. a k/v server) wants to start
agreement on the next command to be appended to Raft's log.
if this server isn't the leader, returns false.
?otherwise start the agreement and return immediately.

there is no guarantee that this command will ever be committed to the Raft log, since the leader
may fail or lose an election. even if the Raft instance has been killed,
this function should return gracefully.

the first return value is the index that the command will appear at
if it's ever committed. the second return value is the current
term. the third return value is true if this server believes it is
the leader.
*/
func (rf *Raft) Start(command []byte) (int, int, bool) {
	// Your code here (2B).

	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state != leader {
		return -1, rf.currentTerm, false
	}

	if rf.leaderId != rf.me {
		// Debug(dError, "S%d accepting command while not the leader. There is a fault in your logic causing rf.state and rf.leaderId to mismatch. term: %d, state: %s leaderId: %d\n", rf.me, rf.currentTerm, rf.state, rf.leaderId)
	}

	entry := logEntry{
		Command: command,
		Index:   len(rf.log),
		Term:    rf.currentTerm,
	}

	rf.log = append(rf.log, entry)
	rf.persist()
	// Debug(dClient, "S%d got a command. entry: %v state: %s log: %v\n", rf.me, entry, rf.state, rf.log)

	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}

		go func(notiferIndex int) {
			select {
			case rf.appendNotifiers[notiferIndex] <- struct{}{}:
			default:
			}

		}(i)
	}

	return entry.Index, entry.Term, true
}

/*
!!the tester doesn't halt goroutines created by Raft after each test,
but it does call the Kill() method. your code can use killed() to
check whether Kill() has been called. the use of atomic avoids the
need for a lock.

!! the issue is that long-running goroutines use memory and may chew
up CPU time, perhaps causing later tests to fail and generating
confusing // debug output.
!! any goroutine with a long-running loop
!! should call killed() to check whether it should stop.
*/
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func startNewTerm(initial bool, rf *Raft) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state == leader {
		// Debug(dInfo, "S%d stopped trying to start a new term as I'm the leader\n")
		return
	}

	rf.currentTerm++
	// DPrintf("starting term %d on peer %d with state %s\n", rf.currentTerm, rf.me, rf.state)
	if initial {
		rf.state = follower
		rf.votesReceived = 0
		rf.votedFor = -1
	} else {
		rf.state = candidate
		rf.votesReceived = 1
		// from lec 6, if a server voted for itself it shouldn't vote for someone else. not a big thing tho
		rf.votedFor = rf.me
	}
	rf.persist()
	// Debug(dTerm, "S%d starting term = %d with state %s\n", rf.me, rf.currentTerm, rf.state)

	rf.leaderId = -1

	if initial {
		return
	}

	for i := 0; i < len(rf.peers); i++ {
		if i != rf.me {
			go repeatRequestVote(i, rf, rf.currentTerm)
		}
	}
}

/*
You'll want to have a separate long-running goroutine that sends
committed log entries in order on the applyCh. It must be separate,
since sending on the applyCh can block; and it must be a single
goroutine, since otherwise it may be hard to ensure that you send log
entries in log order.
*/
func ApplySafeEntries(rf *Raft) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.leaderId == rf.me {
		if rf.state != leader {
			// Debug(dWarn, "S%d state is not leader but leaderId is me\n", rf.me)
		}

		rf.matchIndex[rf.me] = len(rf.log) - 1

		// Debug(dCommit, "S%d leader is about to update commit Index. CommitIndex: %d, matchIndex[]: %v\n", rf.me, rf.commitIndex, rf.matchIndex)

		matches := make([]int, len(rf.matchIndex))
		nCopied := copy(matches, rf.matchIndex)

		if nCopied != len(rf.matchIndex) {
			// Debug(dError, "S%d matchIndex was not correctly coppied\n", rf.me)
		}

		sort.Ints(matches)

		maj := len(rf.peers)/2 + 1

		newCommitIndex := matches[maj-1]

		if newCommitIndex > rf.commitIndex && rf.log[newCommitIndex].Term == rf.currentTerm {
			// Debug(dCommit, "S%d updating commit index from: %d to: %d. log: %v\n", rf.me, rf.commitIndex, newCommitIndex, rf.log)
			rf.commitIndex = newCommitIndex
		}

	}

	// Debug(dCommit, "S%d applying commited entries between lastApplied + 1: %d up to commitIndex: %d. Log: %v\n", rf.me, rf.lastApplied+1, rf.commitIndex, rf.log)

	for i := rf.lastApplied + 1; i <= rf.commitIndex; i++ {
		msg := ApplyMsg{
			Command:      rf.log[i].Command,
			CommandIndex: i,
			CommandValid: true,
		}

		rf.lastApplied++
		// !! should send these msgs in order to the apply chan
		rf.applyCh <- msg
	}
	// Debug(dCommit, "S%d success applying new commited entreis. lastApplied: %d, commitIndex: %d\n", rf.me, rf.lastApplied, rf.commitIndex)

}

// controls append entries to a specfic server associated with the passed notifier channel
// ensures that there will be only one append entry rpc going for the same peer at a time
func appendEntriesController(rf *Raft, peer int, notifier chan struct{}) {
	for {
		if rf.killed() {
			return
		}
		<-notifier
		if rf.killed() {
			return
		}
		rf.mu.Lock()

		if rf.state != leader {
			// Debug(dInfo, "S%d notified to append logs while not the leader. me: %s\n", rf.me, rf)
			rf.mu.Unlock()
			continue
		}
		// ? optimization: might avoid sending these rpc if matchIndex[peer] == nextIndex[peer] - 1
		// if rf.matchIndex[peer] == len(rf.log)-1 {
		// 	// Debug(dInfo, "S%d ignoring append notifier\n", rf.me)
		// 	rf.mu.Unlock()
		// 	return
		// }
		// Debug(dLog, "S%d notified of new entries to replicate to S%d. term: %d, log: %v, nextIndex[%d]: %d\n", rf.me, peer, rf.currentTerm, rf.log, peer, rf.nextIndex[peer])
		args := pb.AppendEntriesArgs{
			Term:         int32(rf.currentTerm),
			LeaderId:     int32(rf.me),
			PrevLogIndex: int32(rf.nextIndex[peer] - 1),
			PrevLogTerm:  int32(rf.log[rf.nextIndex[peer]-1].Term),
			Entries:      MapRaftEntriestoGrpcEntries(rf.log[rf.nextIndex[peer]:]),
			LeaderCommit: int32(rf.commitIndex),
			IsHeartbeat:  false,
		}

		rf.mu.Unlock()

		repeatAppendEntries(peer, rf, &args)
	}
}

func heartbeat(rf *Raft) {
	for {
		if rf.killed() {
			return
		}
		rf.mu.Lock()
		if rf.killed() {
			return
		}

		if rf.leaderId != rf.me || rf.state != leader {
			rf.mu.Unlock()
			break
		}
		// DPrintf("peer %d is sending hearbeats in term %d\n", rf.me, rf.currentTerm)
		// Debug(dLeader, "S%d is sending heartbeats in term %d. state: %s\n", rf.me, rf.currentTerm, rf.state)
		for i := 0; i < len(rf.peers); i++ {
			if i != rf.me {
				appendArgs := pb.AppendEntriesArgs{
					Term:         int32(rf.currentTerm),
					LeaderId:     int32(rf.me),
					PrevLogIndex: int32(len(rf.log) - 1),
					PrevLogTerm:  int32(rf.log[len(rf.log)-1].Term),
					Entries:      MapRaftEntriestoGrpcEntries(make([]logEntry, 0)),
					LeaderCommit: int32(rf.commitIndex),
					IsHeartbeat:  true,
				}

				// !! need to make sure hearbeats are sent immeditally after winning an election
				go repeatAppendEntries(i, rf, &appendArgs)
			}
		}
		rf.mu.Unlock()
		time.Sleep(time.Duration(HBRATE) * time.Millisecond)

	}
}

func timer(rf *Raft) {
	for {
		if rf.killed() {
			return
		}
		rf.mu.Lock()
		if rf.killed() {
			return
		}
		if rf.state != leader && time.Since(rf.lastRpcTime) > time.Duration(rf.electionTimeout)*time.Millisecond {
			// Debug(dTimer, "S%d Timed out. starting a new term\n", rf.me)
			rf.electionTimeout = rand.Intn(MAXTO-MINTO) + MINTO
			// Debug(dTimer, "S%d election time out: %dms\n", rf.me, rf.electionTimeout)
			rf.lastRpcTime = time.Now()
			rf.mu.Unlock()
			startNewTerm(false, rf)
		} else {
			rf.mu.Unlock()
		}

		time.Sleep(time.Duration(rf.electionTimeout) * time.Millisecond)
	}

}

// user of the following funcions should lock before using them
func revertToFollower(rf *Raft, term int, leaderId int) {
	rf.state = follower
	rf.currentTerm = term
	rf.votedFor = -1
	rf.votesReceived = 0
	rf.leaderId = leaderId
	rf.persist()
}
