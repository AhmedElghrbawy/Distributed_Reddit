package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/netip"
	"os"
	"strconv"
	"sync"

	"github.com/ahmedelghrbawy/replicated_db/pkg/raft"
)

type rdbServer struct {
	mu                 sync.Mutex
	shardNum           int
	replicaNum         int
	rf                 *raft.Raft
	applyCh            chan raft.ApplyMsg
	raftPeersAddresses []netip.AddrPort
	myAddress          netip.AddrPort
}

type op struct {
	query string
}

/*
example: ./prog shardNum replicaNum
*/
func main() {
	rdb := &rdbServer{}
	rdb.applyCh = make(chan raft.ApplyMsg)

	parseCommandLineArgs(rdb)

	log.Printf("replicated DB server {shardNum: %d, replicaNum: %v} started\n", rdb.shardNum, rdb.replicaNum)

	parseConfigFile(rdb)

	rdb.rf = raft.Make(rdb.raftPeersAddresses, rdb.replicaNum, rdb.applyCh)

	// go rdb.applyCommands()
}

func (rdb *rdbServer) applyCommands() {
	for {
		msg := <-rdb.applyCh

		op := msg.Command

		log.Printf("got op back %s\n", string(op))

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
