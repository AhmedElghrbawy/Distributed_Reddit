package logger
/*
credits: https://blog.josejg.com/debugging-pretty/
*/
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
var debug int = 1

type logTopic string

const (
	DClient   logTopic = "CLNT"
	DCommit   logTopic = "CMIT"
	DDrop     logTopic = "DROP"
	DError    logTopic = "ERRO"
	DInfo     logTopic = "INFO"
	DLeader   logTopic = "LEAD"
	DLog      logTopic = "LOG1"
	DLog2     logTopic = "LOG2"
	DPersist  logTopic = "PERS"
	DSnap     logTopic = "SNAP"
	DTerm     logTopic = "TERM"
	DTest     logTopic = "TEST"
	DTimer    logTopic = "TIMR"
	DTrace    logTopic = "TRCE"
	DVote     logTopic = "VOTE"
	DElection logTopic = "ELEC"
	DWarn     logTopic = "WARN"
	DNotify   logTopic = "NTFY"
)

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

// Debug(logger.DTimer, "S%d Leader, checking heartbeats", rf.me)

func Debug(topic logTopic, format string, a ...interface{}) {
	if debug >= 1 {
		time := time.Since(debugStart).Microseconds()
		time /= 100
		prefix := fmt.Sprintf("%06d %v ", time, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}
