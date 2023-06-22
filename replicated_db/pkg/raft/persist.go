package raft

import (
	"bytes"
	"log"

	"github.com/ahmedelghrbawy/replicated_db/pkg/labgob"
)

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
