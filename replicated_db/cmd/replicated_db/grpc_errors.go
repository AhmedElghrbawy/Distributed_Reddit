package main

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rdb_grpc_error int

const (
	NOT_THE_LEADER rdb_grpc_error = iota
	SERVER_RESPONSE_TIMEOUT rdb_grpc_error = iota
)

var rdb_grpc_error_map = map[rdb_grpc_error]error{
	NOT_THE_LEADER: status.Error(codes.Unavailable, "not the leader"),
	SERVER_RESPONSE_TIMEOUT: status.Error(codes.DeadlineExceeded, "the server received the request, but timed out waiting for the response to be generated."),
}
