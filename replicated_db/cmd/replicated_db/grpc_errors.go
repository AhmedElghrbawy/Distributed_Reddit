package main

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rdb_grpc_error int

const (
	NOT_THE_LEADER rdb_grpc_error = iota
)

var rdb_grpc_error_map = map[rdb_grpc_error]error{
	NOT_THE_LEADER: status.Error(codes.Unavailable, "not the leader"),
}
