package main

import (
	context "context"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (rdb *rdbServer) GetSubreddit(ctx context.Context, in_subreddit *pb.SubredditInfo) (*pb.Subreddit, error) {
	executer_function := func(){
		query = `SELECT `
	}
}

func (rdb *rdbServer) CreateSubreddit(ctx context.Context, in_subreddit *pb.SubredditInfo) (*pb.Subreddit, error) {
	return nil, nil
}

func (rdb *rdbServer) GetSubreddits(ctx context.Context, in *emptypb.Empty) (*pb.SubredditList, error) {
	return nil, nil
}
