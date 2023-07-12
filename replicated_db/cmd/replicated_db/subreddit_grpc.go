package main

import (
	context "context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) GetSubreddit(ctx context.Context, in_subreddit *pb.Subreddit) (*pb.Subreddit, error) {
	return in_subreddit, nil
}

func (rdb *rdbServer) CreateSubreddit(ctx context.Context, in_subreddit *pb.Subreddit) (*pb.Subreddit, error) {
	return in_subreddit, nil
}

func (rdb *rdbServer) GetSubreddits(ctx context.Context, in *emptypb.Empty) (*pb.SubredditList, error) {
	return nil, nil
}
