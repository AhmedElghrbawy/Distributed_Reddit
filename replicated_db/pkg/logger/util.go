package raft

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

var debugStart time.Time = time.Now()
var debugVerbosity int

// Debugging
var debug int = 0

type logTopic string

const (
	dClient   logTopic = "CLNT"
	dCommit   logTopic = "CMIT"
	dDrop     logTopic = "DROP"
	dError    logTopic = "ERRO"
	dInfo     logTopic = "INFO"
	dLeader   logTopic = "LEAD"
	dLog      logTopic = "LOG1"
	dLog2     logTopic = "LOG2"
	dPersist  logTopic = "PERS"
	dSnap     logTopic = "SNAP"
	dTerm     logTopic = "TERM"
	dTest     logTopic = "TEST"
	dTimer    logTopic = "TIMR"
	dTrace    logTopic = "TRCE"
	dVote     logTopic = "VOTE"
	dElection logTopic = "ELEC"
	dWarn     logTopic = "WARN"
	dNotify   logTopic = "NTFY"
)

// func DPrintf(format string, a ...interface{}) (n int, err error) {
// 	if Debug > 0 {
// 		log.Printf(format, a...)
// 	}
// 	return
// }

func getVerbosity() int {
	v := os.Getenv("VERBOSE")
	level := 0
	if v != "" {
		var err error
		level, err = strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Invalid verbosity %v", v)
		}
	}
	return level
}

func Init() {
	debugVerbosity = getVerbosity()
	// debugStart = time.Now()

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

// Debug(dTimer, "S%d Leader, checking heartbeats", rf.me)

func Debug(topic logTopic, format string, a ...interface{}) {
	if debug >= 1 {
		time := time.Since(debugStart).Microseconds()
		time /= 100
		prefix := fmt.Sprintf("%06d %v ", time, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}
