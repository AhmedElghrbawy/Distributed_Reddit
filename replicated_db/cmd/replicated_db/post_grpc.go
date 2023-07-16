package main

import (
	context "context"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) GetPost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil

}

func (rdb *rdbServer) CreatePost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil
}

func (rdb *rdbServer) GetPosts(ctx context.Context, message_info *pb.MessageInfo) (*pb.PostList, error) {
	return nil, nil
}

func (rdb *rdbServer) Pin(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil
}

func (rdb *rdbServer) Unpin(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil
}

func (rdb *rdbServer) UpVote(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil
}

func (rdb *rdbServer) DownVote(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	return nil, nil
}
