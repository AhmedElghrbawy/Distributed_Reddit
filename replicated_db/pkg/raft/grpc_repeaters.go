package raft

import (
	"github.com/ahmedelghrbawy/replicated_db/pkg/logger"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"
)

func repeatRequestVote(server int, rf *Raft, rpcTerm int) {
	rf.mu.Lock()

	if rf.currentTerm > rpcTerm {
		// ? This go routine is outdated.
		// we are now in a different term than when this go routine was created
		rf.mu.Unlock()
		return
	}
	rvargs := pb.RequestVoteArgs{
		Term:         int32(rf.currentTerm),
		CandidateId:  int32(rf.me),
		LastLogIndex: int32(len(rf.log) - 1),
		LastLogTerm:  int32(rf.log[len(rf.log)-1].Term),
	}

	logger.Debug(logger.DVote, "S%d is sending RV to %d. Term: %d, state: %s, args: %s\n", rf.me, server, rf.currentTerm, rf.state, &rvargs)
	rf.mu.Unlock()

	for {

		rvreply, err := rf.sendRequestVote(server, &rvargs)

		if err != nil {
			continue
		}

		rf.mu.Lock()
		defer rf.mu.Unlock()

		if int(rvreply.Term) > rf.currentTerm {
			revertToFollower(rf, int(rvreply.Term), -1)
			return
		}
		if int(rvargs.Term) < rf.currentTerm {
			// this request is outdated
			return
		}

		if rvreply.VoteGranted {
			rf.votesReceived++
			logger.Debug(logger.DVote, "S%d Got vote from peer %d. Term = %d, votesReceived = %d\n", rf.me, server, rf.currentTerm, rf.votesReceived)
			if rf.votesReceived == len(rf.peers)/2+1 {
				logger.Debug(logger.DLeader, "S%d got a majority for term = %d. Designating as leader...\n", rf.me, rf.currentTerm)

				// send hearbeat immeditaly
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
						go rf.sendAppendEntries(i, &appendArgs)
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

func repeatAppendEntries(server int, rf *Raft, args *pb.AppendEntriesArgs) {

	for {
		rf.mu.Lock()

		if rf.currentTerm > int(args.Term) {
			// This go routine is outdated.
			rf.mu.Unlock()
			return
		}

		if rf.leaderId != rf.me || rf.state != leader {
			logger.Debug(logger.DError, "S%d sending AE while not the leader. term: %d, leaderId: %d\n", rf.me, rf.currentTerm, rf.leaderId)
		}

		rf.mu.Unlock()

		reply, err := rf.sendAppendEntries(server, args)

		if err != nil && args.IsHeartbeat {
			return
		} else if err != nil {
			continue
		}

		rf.mu.Lock()
		if int(reply.Term) > rf.currentTerm {
			revertToFollower(rf, int(reply.Term), -1)
			rf.mu.Unlock()
			return
		}
		if int(args.Term) < rf.currentTerm {
			// this request is outdated
			// logger.Debug(logger.DDrop, "S%d Ignore sending outdated rpc. currentTerm: %d, rpcTerm: %d\n", rf.me, rf.currentTerm, args.Term)
			rf.mu.Unlock()
			return
		}

		if args.IsHeartbeat {
			if !reply.Success {
				// logger.Debug(logger.DNotify, "S%d detected mismatching log for peer %d while sending heartbeats. me: %s, args: %s\n", rf.me, server, rf, args)
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
				// logger.Debug(logger.DLog2, "S%d follower's log is too short\n", rf.me)
				rf.nextIndex[server] = int(reply.XLen)
			} else {
				xTermLastIndex := -1

				for i, entry := range rf.log {
					if entry.Term == int(reply.XTerm) {
						xTermLastIndex = i
					} else if entry.Term > int(reply.XTerm) {
						break
					}
				}

				if xTermLastIndex == -1 {
					// logger.Debug(logger.DLog2, "S%d no entry with xterm\n", rf.me)
					rf.nextIndex[server] = int(reply.XIndex)
				} else {
					rf.nextIndex[server] = xTermLastIndex
				}

			}

			args.Entries = MapRaftEntriestoGrpcEntries(rf.log[rf.nextIndex[server]:])
			args.PrevLogIndex = int32(rf.nextIndex[server] - 1)
			args.PrevLogTerm = int32(rf.log[args.PrevLogIndex].Term)
			args.LeaderCommit = int32(rf.commitIndex)
			// args.Term = rf.currentTerm
			logger.Debug(logger.DLog, "S%d AE request to %d failed due to consistency check. Decremented nextIndex. me: %s, args: %s\n", rf.me, server, rf, args)
			rf.mu.Unlock()
			continue
		}

		rf.nextIndex[server] = rf.nextIndex[server] + len(args.Entries)
		rf.matchIndex[server] = rf.nextIndex[server] - 1
		logger.Debug(logger.DLog, "S%d replicated entries: %v to S%d. me: %s, args: %s\n", rf.me, args.Entries, server, rf, args)
		go applySafeEntries(rf)
		rf.mu.Unlock()
		return

	}

}
