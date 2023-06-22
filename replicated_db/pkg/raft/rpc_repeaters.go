package raft

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
