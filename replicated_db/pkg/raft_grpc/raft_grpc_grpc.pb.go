// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.3
// source: raft_grpc.proto

package raft_grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	RaftGRPC_RequestVote_FullMethodName   = "/raft_grpc.RaftGRPC/RequestVote"
	RaftGRPC_AppendEntries_FullMethodName = "/raft_grpc.RaftGRPC/AppendEntries"
)

// RaftGRPCClient is the client API for RaftGRPC service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftGRPCClient interface {
	RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error)
	AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error)
}

type raftGRPCClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftGRPCClient(cc grpc.ClientConnInterface) RaftGRPCClient {
	return &raftGRPCClient{cc}
}

func (c *raftGRPCClient) RequestVote(ctx context.Context, in *RequestVoteArgs, opts ...grpc.CallOption) (*RequestVoteReply, error) {
	out := new(RequestVoteReply)
	err := c.cc.Invoke(ctx, RaftGRPC_RequestVote_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftGRPCClient) AppendEntries(ctx context.Context, in *AppendEntriesArgs, opts ...grpc.CallOption) (*AppendEntriesReply, error) {
	out := new(AppendEntriesReply)
	err := c.cc.Invoke(ctx, RaftGRPC_AppendEntries_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftGRPCServer is the server API for RaftGRPC service.
// All implementations must embed UnimplementedRaftGRPCServer
// for forward compatibility
type RaftGRPCServer interface {
	RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error)
	AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error)
	mustEmbedUnimplementedRaftGRPCServer()
}

// UnimplementedRaftGRPCServer must be embedded to have forward compatible implementations.
type UnimplementedRaftGRPCServer struct {
}

func (UnimplementedRaftGRPCServer) RequestVote(context.Context, *RequestVoteArgs) (*RequestVoteReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (UnimplementedRaftGRPCServer) AppendEntries(context.Context, *AppendEntriesArgs) (*AppendEntriesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftGRPCServer) mustEmbedUnimplementedRaftGRPCServer() {}

// UnsafeRaftGRPCServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftGRPCServer will
// result in compilation errors.
type UnsafeRaftGRPCServer interface {
	mustEmbedUnimplementedRaftGRPCServer()
}

func RegisterRaftGRPCServer(s grpc.ServiceRegistrar, srv RaftGRPCServer) {
	s.RegisterService(&RaftGRPC_ServiceDesc, srv)
}

func _RaftGRPC_RequestVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestVoteArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftGRPCServer).RequestVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftGRPC_RequestVote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftGRPCServer).RequestVote(ctx, req.(*RequestVoteArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _RaftGRPC_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftGRPCServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftGRPC_AppendEntries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftGRPCServer).AppendEntries(ctx, req.(*AppendEntriesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftGRPC_ServiceDesc is the grpc.ServiceDesc for RaftGRPC service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftGRPC_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "raft_grpc.RaftGRPC",
	HandlerType: (*RaftGRPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestVote",
			Handler:    _RaftGRPC_RequestVote_Handler,
		},
		{
			MethodName: "AppendEntries",
			Handler:    _RaftGRPC_AppendEntries_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "raft_grpc.proto",
}
