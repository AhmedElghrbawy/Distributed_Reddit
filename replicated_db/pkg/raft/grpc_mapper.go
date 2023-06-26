package raft

import pb "github.com/ahmedelghrbawy/replicated_db/pkg/raft_grpc"

func MapGrpcEntriesToRaftEntries(src []*pb.LogEntry) []logEntry {
	result := make([]logEntry, 0)

	for _, entry := range src {
		result = append(result, logEntry{
			Command: entry.Command,
			Term:    int(entry.Term),
			Index:   int(entry.Index),
		})
	}

	return result
}


func MapRaftEntriestoGrpcEntries(src []logEntry) []*pb.LogEntry {
	result := make([]*pb.LogEntry, 0)

	for _, entry := range src {
		result = append(result, &pb.LogEntry{
			Command: entry.Command,
			Term:    int32(entry.Term),
			Index:   int32(entry.Index),
		})
	}

	return result
}