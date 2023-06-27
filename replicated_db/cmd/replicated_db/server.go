package main

import (
	"log"
	"net/netip"
	"os"
	"strconv"
	"sync"

	"github.com/ahmedelghrbawy/replicated_db/pkg/raft"
)

type rdbServer struct {
	mu            sync.Mutex
	me            int
	rf            *raft.Raft
	applyCh       chan raft.ApplyMsg
	peerAddresses []netip.AddrPort
}

/*
command line arguments

	{
		nPeers: int                // number of peers
		pAddI:  ip:port           // address of the i'th peer
		me:     int               // index of this server in peers list
	}

example: ./prog npeers pAdd0 pAdd1 ... pAddn me
*/
func main() {
	rdbServer := &rdbServer{}
	rdbServer.applyCh = make(chan raft.ApplyMsg)

	parseCommandLineArgs(rdbServer)

	log.Printf("replicated DB server started with state {me: %d, peers: %v}\n", rdbServer.me, rdbServer.peerAddresses)

	rdbServer.rf = raft.Make(rdbServer.peerAddresses, rdbServer.me, rdbServer.applyCh)

}

func parseCommandLineArgs(rdbServer *rdbServer) {
	args := os.Args[1:]
	log.Printf("command line arguments: %v\n", args)

	// npeers
	npeers, err := strconv.ParseInt(args[0], 10, 32)

	if err != nil {
		log.Fatalf("failed to parse command line argument npeers, %v\n", err)
	}

	// peers
	peers := make([]netip.AddrPort, 0)
	for i := 0; i < int(npeers); i++ {
		add, err := netip.ParseAddrPort(args[i+1])
		if err != nil {
			log.Fatalf("couldn't parse address from command line arguments arg: %s, %v", args[i+1], err)
		}

		peers = append(peers, add)
	}

	rdbServer.peerAddresses = peers

	// me
	me, err := strconv.ParseInt(args[npeers+1], 10, 32)
	if err != nil {
		log.Fatalf("failed to parse command line argument me, %v\n", err)
	}

	rdbServer.me = int(me)

}
