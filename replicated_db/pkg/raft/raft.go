package raft

/*
	this is an outline of the API that raft must expose to the application/service using it.

	create a new Raft server:
		rf = Make(...)

	start agreement on a new log entry:
		rf.Start(command []byte) (index, term, isleader)

	ask Raft for its current term, and whether it thinks it is leader:
		rf.GetState() (term, isLeader)

	ApplyMsg
		each time a new entry is committed to the log, each Raft peer
		should send an ApplyMsg to the application.
*/

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"sort"
	"sync"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/logger"
	"github.com/ahmedelghrbawy/replicated_db/pkg/persister"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
as each Raft peer becomes aware that successive log entries are
committed, the peer should send an ApplyMsg to the service on the same server,
via the applyCh passed to Make().
*/
type ApplyMsg struct {
	Command      []byte
	CommandIndex int
}

/*
the application should encode its command and send it as a []byte for raft to store it in the logs.
this helps make raft more decoupled from the application running on top of it, and it makes sending log entries in gRPC easier.
*/
type logEntry struct {
	Command []byte
	Term    int
	Index   int
}

func (e *logEntry) String() string {
	return fmt.Sprintf("{Term: %d, Index: %d, Command: %s}", e.Term, e.Index, string(e.Command[:]))
}

type state string

const (
	follower  state = "follower"
	candidate state = "candidate"
	leader    state = "leader"
)
const (
	MINTO  int = 400
	MAXTO  int = 750
	HBRATE int = 40
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex           // Lock to protect shared access to this peer's state
	peers     []pb.RaftGRPCClient  // each index contains a gRPC client for the corresponding peer
	persister *persister.Persister // Object to hold this peer's persisted state
	me        int                  // this peer's index into peers[]

	pb.UnimplementedRaftGRPCServer

	// raft state variables following Figure 2 in the paper

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
the service wants to create a Raft server.
the IP's of all Raft servers (including this one) are in peerAddresses[].
this server's IP is peers[me].
all the servers' peers[] arrays have the same order.
persister is a place for this server to save its persistent state, and also initially holds the most recent saved state, if any.
applyCh is a channel on which the service expects Raft to send ApplyMsg messages.
Make() must return quickly, so it should start goroutines for any long-running work.
*/
func Make(peerAddresses []netip.AddrPort, me int, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.persister = persister.MakePersister()
	rf.me = me
	rf.applyCh = applyCh
	rf.appendNotifiers = make([]chan struct{}, 0)
	rf.currentTerm = -1
	rf.votedFor = -1
	rf.lastRpcTime = time.Now()
	rf.log = []logEntry{{
		Command: make([]byte, 0),
		Index:   0,
		Term:    -1,
	}}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.peers = make([]pb.RaftGRPCClient, 0)

	logger.Init()

	lis, err := net.Listen("tcp", peerAddresses[me].String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpc_server := grpc.NewServer()
	pb.RegisterRaftGRPCServer(grpc_server, rf)

	go func() {
		if err := grpc_server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}

	}()

	for i, peerAddress := range peerAddresses {
		if i == me {
			rf.peers = append(rf.peers, nil)
			continue
		}
		// TODO: need to clean up these connections somewhere, maybe send a cancellation signal from the application layer
		conn, err := grpc.Dial(peerAddress.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("unable to initialize a client : %v", err)
		}
		// defer conn.Close()

		client := pb.NewRaftGRPCClient(conn)

		rf.peers = append(rf.peers, client)
	}

	rf.nextIndex = make([]int, len(rf.peers))

	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = 1
	}

	rf.matchIndex = make([]int, len(rf.peers))
	rf.electionTimeout = rand.Intn(MAXTO-MINTO) + MINTO

	// initialize from state persisted before a crash
	rf.readPersist(rf.persister.ReadRaftState())

	// each peer will have a long running notifier go routine that waits to be notified of a new command or a won ellection and starts sending append entries to this peer
	// this helps avoiding inconsistencies
	for i := 0; i < len(rf.peers); i++ {
		rf.appendNotifiers = append(rf.appendNotifiers, make(chan struct{}))
		if i == rf.me {
			continue
		}
		go appendEntriesController(rf, i, rf.appendNotifiers[i])
	}

	logger.Debug(logger.DInfo, "S%d Starting to operate...\n", rf.me)
	// logger.Debug(logger.DTimer, "S%d election time out: %dms\n", rf.me, rf.electionTimeout)

	startNewTerm(true, rf)
	go timer(rf)

	return rf
}

// returns currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// logger.Debug(logger.DInfo, "S%d Getting state... current term = %d, leader = %d state = %s\n", rf.me, rf.currentTerm, rf.leaderId, rf.state)
	return rf.currentTerm, rf.leaderId == rf.me

}

/*
the service using Raft wants to start
agreement on the next command to be appended to Raft's log.
if this server isn't the leader, returns false.
otherwise start the agreement and return immediately.

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

	entry := logEntry{
		Command: command,
		Index:   len(rf.log),
		Term:    rf.currentTerm,
	}

	rf.log = append(rf.log, entry)
	rf.persist()
	logger.Debug(logger.DClient, "S%d got a command. entry: %s state: %s log: %v\n", rf.me, entry.String(), rf.state, rf.log)

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

func startNewTerm(initial bool, rf *Raft) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state == leader {
		// logger.Debug(logger.DInfo, "S%d stopped trying to start a new term as I'm the leader\n")
		return
	}

	rf.currentTerm++
	if initial {
		rf.state = follower
		rf.votesReceived = 0
		rf.votedFor = -1
	} else {
		rf.state = candidate
		rf.votesReceived = 1
		rf.votedFor = rf.me
	}
	rf.persist()
	logger.Debug(logger.DTerm, "S%d starting term = %d with state %s\n", rf.me, rf.currentTerm, rf.state)

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
a separate long-running goroutine that sends
committed log entries in order on the applyCh. It must be separate,
since sending on the applyCh can block; and it must be a single
goroutine, since otherwise it may be hard to ensure that you send log
entries in log order.
*/
func applySafeEntries(rf *Raft) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.leaderId == rf.me {

		rf.matchIndex[rf.me] = len(rf.log) - 1

		logger.Debug(logger.DCommit, "S%d leader is about to update commit Index. CommitIndex: %d, matchIndex[]: %v\n", rf.me, rf.commitIndex, rf.matchIndex)

		matches := make([]int, len(rf.matchIndex))
		copy(matches, rf.matchIndex)

		sort.Ints(matches)

		maj := len(rf.peers)/2 + 1

		newCommitIndex := matches[maj-1]

		if newCommitIndex > rf.commitIndex && rf.log[newCommitIndex].Term == rf.currentTerm {
			logger.Debug(logger.DCommit, "S%d updating commit index from: %d to: %d. log: %v\n", rf.me, rf.commitIndex, newCommitIndex, rf.log)
			rf.commitIndex = newCommitIndex
		}

	}

	logger.Debug(logger.DCommit, "S%d applying commited entries between lastApplied + 1: %d up to commitIndex: %d. Log: %v\n", rf.me, rf.lastApplied+1, rf.commitIndex, rf.log)

	for i := rf.lastApplied + 1; i <= rf.commitIndex; i++ {
		msg := ApplyMsg{
			Command:      rf.log[i].Command,
			CommandIndex: i,
		}

		rf.lastApplied++
		// !! should send these msgs in order to the apply chan
		rf.applyCh <- msg
	}
	logger.Debug(logger.DCommit, "S%d success applying new commited entreis. lastApplied: %d, commitIndex: %d\n", rf.me, rf.lastApplied, rf.commitIndex)

}

// controls append entries to a specfic server associated with the passed notifier channel
// ensures that there will be only one path of execution responsible for sending
// append entry rpc going for the same peer at a time
func appendEntriesController(rf *Raft, peer int, notifier chan struct{}) {
	for {
		<-notifier

		rf.mu.Lock()

		if rf.state != leader {
			// logger.Debug(logger.DInfo, "S%d notified to append logs while not the leader. me: %s\n", rf.me, rf)
			rf.mu.Unlock()
			continue
		}

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

		rf.mu.Lock()

		if rf.leaderId != rf.me || rf.state != leader {
			rf.mu.Unlock()
			break
		}
		// logger.Debug(logger.DLeader, "S%d is sending heartbeats in term %d. state: %s\n", rf.me, rf.currentTerm, rf.state)
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

		rf.mu.Lock()

		if rf.state != leader && time.Since(rf.lastRpcTime) > time.Duration(rf.electionTimeout)*time.Millisecond {
			logger.Debug(logger.DTimer, "S%d Timed out. starting a new term\n", rf.me)
			rf.electionTimeout = rand.Intn(MAXTO-MINTO) + MINTO
			// logger.Debug(logger.DTimer, "S%d election time out: %dms\n", rf.me, rf.electionTimeout)
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
