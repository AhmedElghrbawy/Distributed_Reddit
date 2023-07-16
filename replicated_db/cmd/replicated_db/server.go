package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ahmedelghrbawy/replicated_db/pkg/raft"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

type rdbServer struct {
	shardNum           int
	replicaNum         int
	dbConnectionStr    string
	rf                 *raft.Raft
	mu                 sync.Mutex
	applyCh            chan raft.ApplyMsg
	raftPeersAddresses []netip.AddrPort
	myAddress          netip.AddrPort

	pb.UnimplementedSubredditGRPCServer

	// needs to be locked
	replyMap map[string]*replyInfo
}

type Op struct {
	Executer Executer
	Id                string
}

type CommandNotExecutedError struct{}

func (e *CommandNotExecutedError) Error() string {
	return "command was not executed"
}

type replyInfo struct {
	id string
	// ready  bool
	result interface{}
	err    error
	ch     (chan struct{})
}

const (
	db_host = "localhost"
	db_port = 5432
	db_user = "postgres"
)

/*
example: ./prog shardNum replicaNum
*/
func main() {
	// initialize rdb
	rdb := &rdbServer{}
	rdb.applyCh = make(chan raft.ApplyMsg)
	rdb.replyMap = make(map[string]*replyInfo)

	parseCommandLineArgs(rdb)
	log.Printf("replicated DB server {shardNum: %d, replicaNum: %v} started\n", rdb.shardNum, rdb.replicaNum)
	parseConfigFile(rdb)

	db_name := fmt.Sprintf("S%d_R%d", rdb.shardNum, rdb.replicaNum)
	db_pass := os.Getenv("PSQL_PASS")
	rdb.dbConnectionStr = fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	db_host, db_port, db_user, db_pass, db_name)
	
	rdb.rf = raft.Make(rdb.raftPeersAddresses, rdb.replicaNum, rdb.applyCh)

	// register executers
	gob.Register(&GetSubredditExecuter{})
	gob.Register(&CreateSubredditExecuter{})
	gob.Register(&GetSubredditsHandlesExecuter{})


	go rdb.applyCommands()

	lis, err := net.Listen("tcp", rdb.myAddress.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpc_server := grpc.NewServer()
	pb.RegisterSubredditGRPCServer(grpc_server, rdb)

	if err := grpc_server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Printf("shard: %d, replica %d listening on: %s\n", rdb.shardNum, rdb.replicaNum, rdb.myAddress.String())
}

func (rdb *rdbServer) applyCommands() {
	for {
		msg := <-rdb.applyCh

		log.Println("got a msg back from raft")

		opBytes := msg.Command
		dec := gob.NewDecoder(bytes.NewBuffer(opBytes))

		var op Op

		err := dec.Decode(&op)
		if err != nil {
			log.Fatal("decode error 1:", err)
		}
		fmt.Printf("decoded op returned from raft %v\n" ,op)

		result, err := op.Executer.Execute(rdb)


		
		replyInfo, replyExists  := rdb.replyMap[op.Id]

		if replyExists {
			replyInfo.result = result
			replyInfo.err = err
			replyInfo.ch <- struct{}{}
			select {
			case replyInfo.ch <- struct{}{}:
			case <-time.After(time.Second): // ? magic number
			}
		}
		
		// log.Printf("got op back %s\n", string(op))

	}
}

func parseCommandLineArgs(rdb *rdbServer) {

	args := os.Args[1:]
	log.Printf("command line arguments: %v\n", args)

	// shardNum
	shardNum, err := strconv.ParseInt(args[0], 10, 32)

	if err != nil {
		log.Fatalf("failed to parse command line argument npeers, %v\n", err)
	}
	rdb.shardNum = int(shardNum)

	// replicaNum
	replicaNum, err := strconv.ParseInt(args[1], 10, 32)

	if err != nil {
		log.Fatalf("failed to parse command line argument npeers, %v\n", err)
	}
	rdb.replicaNum = int(replicaNum)

}

type Config struct {
	NumberOfShards        int        `json:"number_of_shards"`
	Number_of_replicas    int        `json:"number_of_replicas"`
	Shard_rdb_servers_ips [][]string `json:"shard_rdb_servers_ips"`
	Shard_raft_peers_ips  [][]string `json:"shard_raft_peers_ips"`
}

func parseConfigFile(rdb *rdbServer) {
	file, err := ioutil.ReadFile("../config.json")
	if err != nil {
		log.Fatalf("couldn't read config file, %v\n", err)
	}

	var config Config
	json.Unmarshal(file, &config)

	myAddress, err := netip.ParseAddrPort(config.Shard_rdb_servers_ips[rdb.shardNum][rdb.replicaNum])
	if err != nil {
		log.Fatalf("couldn't parse myAddress from config: %s, %v", config.Shard_rdb_servers_ips[rdb.shardNum][rdb.replicaNum], err)
	}
	rdb.myAddress = myAddress

	raftPeersAddresses := make([]netip.AddrPort, 0)

	for i := 0; i < config.Number_of_replicas; i++ {
		add, err := netip.ParseAddrPort(config.Shard_raft_peers_ips[rdb.shardNum][i])
		if err != nil {
			log.Fatalf("couldn't parse raft peer address from config: %s, %v", config.Shard_raft_peers_ips[rdb.shardNum][i], err)
		}
		raftPeersAddresses = append(raftPeersAddresses, add)
	}

	rdb.raftPeersAddresses = raftPeersAddresses

}
