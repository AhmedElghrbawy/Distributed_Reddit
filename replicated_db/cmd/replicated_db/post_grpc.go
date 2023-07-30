package main

import (
	context "context"
	"errors"
	"time"

	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
)

func (rdb *rdbServer) GetPost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &GetPostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return replyInfo.result.(*PostDTO).mapToProto(), nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}

}

func (rdb *rdbServer) CreatePost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {
	op := Op{
		Executer: &CreatePostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return in_post_info.Post, nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) GetPosts(ctx context.Context, message_info *pb.MessageInfo) (*pb.PostList, error) {
	op := Op{
		Executer: &GetPostsExecuter{},
		Id:       message_info.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			postsDtos := replyInfo.result.(*[]PostDTO)
			postsProtos := make([]*pb.Post, 0)

			for _, postDto := range *postsDtos {
				postsProtos = append(postsProtos, postDto.mapToProto())
			}
			return &pb.PostList{Posts: postsProtos}, nil

		} else {
			return nil, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}

func (rdb *rdbServer) UpdatePost(ctx context.Context, in_post_info *pb.PostInfo) (*pb.Post, error) {

	if len(in_post_info.UpdatedColumns) == 0 {
		return nil, errors.New("0 columns passed to update post")
	}

	op := Op{
		Executer: &UpdatePostExecuter{In_post_info: in_post_info},
		Id:       in_post_info.MessageInfo.Id,
	}

	submited, replyInfo := rdb.submitOperationToRaft(op)

	if !submited {
		return nil, rdb_grpc_error_map[NOT_THE_LEADER]
	}

	select {
	case <-replyInfo.ch:
		if replyInfo.err == nil {
			return replyInfo.result.(*PostDTO).mapToProto(), nil
		} else {
			return &pb.Post{}, replyInfo.err
		}
	case <-time.After(time.Second): // ? magic number
		return nil, errors.New("timed out")
	}
}
