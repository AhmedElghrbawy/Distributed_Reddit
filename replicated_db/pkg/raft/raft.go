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
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/labgob"

	"github.com/ahmedelghrbawy/replicated_db/pkg/labrpc"
	"github.com/ahmedelghrbawy/replicated_db/pkg/persister"
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
	Command      interface{}
	CommandIndex int
}

type logEntry struct {
	Command interface{}
	Term    int
	Index   int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex           // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd  // RPC end points of all peers
	persister *persister.Persister // Object to hold this peer's persisted state
	me        int                  // this peer's index into peers[]
	dead      int32                // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

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

// warning: "term changed even though there were no failures". this warnning showed when running 2a tests
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

type rpcType string

const (
	APPENT rpcType = "AE"
	REQVOT rpcType = "RV"
	HERTBT rpcType = "HB"
)

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	// 2A
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Debug(dInfo, "S%d Getting state... current term = %d, leader = %d state = %s\n", rf.me, rf.currentTerm, rf.leaderId, rf.state)
	return rf.currentTerm, rf.leaderId == rf.me

}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)

	w := new(bytes.Buffer)
	enc := labgob.NewEncoder(w)

	err := enc.Encode(rf.currentTerm)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	err = enc.Encode(rf.votedFor)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	err = enc.Encode(rf.log)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	data := w.Bytes()
	rf.persister.SaveRaftState(data)

}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }

	if rf.persister.RaftStateSize() == 0 {
		// Debug(dPersist, "S%d raft state size=0. no presisted data to be read\n", rf.me)
		return
	}
	r := bytes.NewBuffer(data)
	dec := labgob.NewDecoder(r)

	var currentTerm int
	var votedFor int
	var l []logEntry

	err := dec.Decode(&currentTerm)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	err = dec.Decode(&votedFor)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	err = dec.Decode(&l)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	rf.currentTerm = currentTerm
	rf.votedFor = votedFor
	rf.log = l

}

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
	return fmt.Sprintf("{Term: %d, candidateId: %d, LastLogIndex: %d, lastLogTerm: %d}       ", args.Term, args.CandidateId, args.LastLogIndex, args.LastLogTerm)
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
	return fmt.Sprintf("{Term: %d, LeaderId: %d, prevLogIndex: %d, preLogTerm: %d, Entries: %v, leaderCommit: %d, IsHeartBeat: %t}         ", args.Term, args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, args.Entries, args.LeaderCommit, args.IsHertbeat)
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	XTerm   int
	XIndex  int
	XLen    int
}

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
// ? this methods sends a request vote rpc
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	// Debug(dLeader, "S%d sending append entries to: %d, args: %s\n", rf.me, server, args)
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
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
func (rf *Raft) Start(command interface{}) (int, int, bool) {
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

/*
the service or tester wants to create a Raft server.
the ports of all the Raft servers (including this one) are in peers[].
this server's port is peers[me].
all the servers' peers[] arrays have the same order.
persister is a place for this server to save its persistent state, and also initially holds the most recent saved state, if any.
applyCh is a channel on which the tester or service expects Raft to send ApplyMsg messages.
Make() must return quickly, so it should start goroutines for any long-running work.
*/
func Make(peers []*labrpc.ClientEnd, me int,
	persister *persister.Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.applyCh = applyCh
	rf.appendNotifiers = make([]chan struct{}, 0)
	rf.currentTerm = -1
	rf.votedFor = -1
	rf.lastRpcTime = time.Now()
	rf.log = []logEntry{logEntry{
		Command: "Dummy command",
		Index:   0,
		Term:    -1,
	}}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))

	for i := 0; i < len(peers); i++ {
		rf.nextIndex[i] = 1
	}

	rf.matchIndex = make([]int, len(rf.peers))
	rf.electionTimeout = rand.Intn(MAXTO-MINTO) + MINTO

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// each peer will have a long running notifier go routine that waits to be notified of a new command or a won ellection and starts sending append entries to this peer
	// this helps avoiding inconsistencies
	for i := 0; i < len(peers); i++ {
		rf.appendNotifiers = append(rf.appendNotifiers, make(chan struct{}))
		if i == rf.me {
			continue
		}
		go appendLogsController(rf, i, rf.appendNotifiers[i])
	}

	// DPrintf("Starting peer %d\n", rf.me)
	// Debug(dInfo, "S%d Starting to operate...\n", rf.me)
	// Debug(dTimer, "S%d election time out: %dms\n", rf.me, rf.electionTimeout)
	// Your initialization code here (2A, 2B, 2C).

	startNewTerm(true, rf)
	go timer(rf)

	return rf
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

func repeatRequestVote(server int, rf *Raft, rpcTerm int) {
	rf.mu.Lock()

	if rf.currentTerm > rpcTerm {
		// ? This go routine is outdated. we are now in a different term than when this go routine was created check Rule number 5 in "advice about locks"
		rf.mu.Unlock()
		return
	}
	rvargs := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: len(rf.log) - 1,
		LastLogTerm:  rf.log[len(rf.log)-1].Term,
	}
	rvreply := RequestVoteReply{}

	// Debug(dVote, "S%d is sending RV to %d. Term: %d, state: %s, args: %s\n", rf.me, server, rf.currentTerm, rf.state, rvargs)
	rf.mu.Unlock()

	for {
		if rf.killed() {
			return
		}
		res := rf.sendRequestVote(server, &rvargs, &rvreply)

		if rf.killed() {
			return
		}
		if !res {
			continue
		}

		rf.mu.Lock()
		defer rf.mu.Unlock()

		if rvreply.Term > rf.currentTerm {
			revertToFollower(rf, rvreply.Term, -1)
			return
		}
		if rvargs.Term < rf.currentTerm {
			// this request is outdated
			return
		}

		if rvreply.VoteGranted {
			rf.votesReceived++
			// DPrintf("Vote is granted for peer %d from peer %d in term %d. votesReceived = %d\n", rf.me, server, rf.currentTerm, rf.votesReceived)
			// Debug(dVote, "S%d Got vote from peer %d. Term = %d, votesReceived = %d\n", rf.me, server, rf.currentTerm, rf.votesReceived)
			if rf.votesReceived == len(rf.peers)/2+1 {
				// DPrintf("Peer %d got a majority for term %d\n", rf.me, rf.currentTerm)
				// Debug(dLeader, "S%d got a majority for term = %d. Designating as leader...\n", rf.me, rf.currentTerm)

				// send hearbeat immeditaly
				for i := 0; i < len(rf.peers); i++ {
					if i != rf.me {
						appendArgs := AppendEntriesArgs{
							Term:         rf.currentTerm,
							LeaderId:     rf.me,
							PrevLogIndex: len(rf.log) - 1,
							PrevLogTerm:  rf.log[len(rf.log)-1].Term,
							Entries:      make([]logEntry, 0),
							LeaderCommit: rf.commitIndex,
							IsHertbeat:   true,
						}

						appendReply := AppendEntriesReply{}

						// !! need to make sure hearbeats are sent immeditally after winning an election
						go rf.sendAppendEntries(i, &appendArgs, &appendReply)
					}
				}
				go heartbeat(rf)
				rf.state = leader
				rf.leaderId = rf.me

				rf.nextIndex = make([]int, len(rf.peers))
				rf.matchIndex = make([]int, len(rf.peers))
				for i := 0; i < len(rf.peers); i++ {
					if i == rf.me {
						rf.matchIndex[i] = len(rf.log) - 1
						continue
					}
					rf.nextIndex[i] = len(rf.log)
					rf.matchIndex[i] = 0
					go func(i int) {
						select {
						case rf.appendNotifiers[i] <- struct{}{}:
						default:
						}
					}(i)
				}

			}
		}

		return
	}
}

func repeatAppendEntries(server int, rf *Raft, args *AppendEntriesArgs, reply *AppendEntriesReply) {

	for {
		if rf.killed() {
			return
		}
		rf.mu.Lock()
		if rf.killed() {
			return
		}

		// should also check if we are the leader for current term, maybe not
		if rf.currentTerm > args.Term {
			// This go routine is outdated.
			rf.mu.Unlock()
			return
		}

		if rf.leaderId != rf.me || rf.state != leader {
			// Debug(dError, "S%d sending AE while not the leader. term: %d, leaderId: %d\n", rf.me, rf.currentTerm, rf.leaderId)
		}

		rf.mu.Unlock()

		res := rf.sendAppendEntries(server, args, reply)
		if rf.killed() {
			return
		}

		if !res && args.IsHertbeat {
			return
		} else if !res {
			continue
		}

		rf.mu.Lock()
		if reply.Term > rf.currentTerm {
			revertToFollower(rf, reply.Term, -1)
			rf.mu.Unlock()
			return
		}
		if args.Term < rf.currentTerm {
			// this request is outdated
			// Debug(dDrop, "S%d Ignore sending outdated rpc. currentTerm: %d, rpcTerm: %d\n", rf.me, rf.currentTerm, args.Term)
			rf.mu.Unlock()
			return
		}

		if args.IsHertbeat {
			if !reply.Success {
				// Debug(dNotify, "S%d detected mismatching log for peer %d while sending heartbeats. me: %s, args: %s\n", rf.me, server, rf, args)
				go func(server int) {
					select {
					case rf.appendNotifiers[server] <- struct{}{}:
					default:
					}
				}(server)
			}
			rf.mu.Unlock()
			return
		}

		if !reply.Success {
			// rf.nextIndex[server]--

			if reply.XTerm == -1 {
				// Debug(dLog2, "S%d follower's log is too short\n", rf.me)
				rf.nextIndex[server] = reply.XLen
			} else {
				xTermLastIndex := -1

				for i, entry := range rf.log {
					if entry.Term == reply.XTerm {
						xTermLastIndex = i
					} else if entry.Term > reply.XTerm {
						break
					}
				}

				if xTermLastIndex == -1 {
					// Debug(dLog2, "S%d no entry with xterm\n", rf.me)
					rf.nextIndex[server] = reply.XIndex
				} else {
					rf.nextIndex[server] = xTermLastIndex
				}

			}

			args.Entries = rf.log[rf.nextIndex[server]:]
			args.PrevLogIndex = rf.nextIndex[server] - 1
			args.PrevLogTerm = rf.log[args.PrevLogIndex].Term
			args.LeaderCommit = rf.commitIndex
			// args.Term = rf.currentTerm
			// Debug(dLog, "S%d AE request to %d failed due to consistency check. Decremented nextIndex. me: %s, args: %s\n", rf.me, server, rf, args)
			rf.mu.Unlock()
			continue
		}

		rf.nextIndex[server] = rf.nextIndex[server] + len(args.Entries)
		rf.matchIndex[server] = rf.nextIndex[server] - 1
		// Debug(dLog, "S%d replicated entries: %v to S%d. me: %s, args: %s\n", rf.me, args.Entries, server, rf, args)
		go ApplySafeEntries(rf)
		rf.mu.Unlock()
		return

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
func appendLogsController(rf *Raft, peer int, notifier chan struct{}) {
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
		args := AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: rf.nextIndex[peer] - 1,
			PrevLogTerm:  rf.log[rf.nextIndex[peer]-1].Term,
			Entries:      rf.log[rf.nextIndex[peer]:],
			LeaderCommit: rf.commitIndex,
			IsHertbeat:   false,
		}

		reply := AppendEntriesReply{}
		rf.mu.Unlock()

		repeatAppendEntries(peer, rf, &args, &reply)
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
				appendArgs := AppendEntriesArgs{
					Term:         rf.currentTerm,
					LeaderId:     rf.me,
					PrevLogIndex: len(rf.log) - 1,
					PrevLogTerm:  rf.log[len(rf.log)-1].Term,
					Entries:      make([]logEntry, 0),
					LeaderCommit: rf.commitIndex,
					IsHertbeat:   true,
				}

				appendReply := AppendEntriesReply{}

				// !! need to make sure hearbeats are sent immeditally after winning an election
				go repeatAppendEntries(i, rf, &appendArgs, &appendReply)
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

// user of the following funcion should lock before using them
func revertToFollower(rf *Raft, term int, leaderId int) {
	rf.state = follower
	rf.currentTerm = term
	rf.votedFor = -1
	rf.votesReceived = 0
	rf.leaderId = leaderId
	rf.persist()
}
