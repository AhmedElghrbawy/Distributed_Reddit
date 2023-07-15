package main

import (
	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentDTO struct {
	model.SubredditComments
}

func (comment_dto *CommentDTO) mapToProto() *pb.Comment {
	proto := &pb.Comment{
		Id:              comment_dto.ID.String(),
		Content:         comment_dto.Content,
		NumberOfVotes:   comment_dto.NumberOfVotes,
		OwnerHandle:     comment_dto.OwnerHandle,
		PostId:          comment_dto.PostID.String(),
	}

	if comment_dto.Image == nil {
		proto.Image = []byte("")
	} else {
		proto.Image = *comment_dto.Image
	}

	if comment_dto.ParentCommentID == nil {
		proto.ParentCommentId = ""
	} else {
		proto.ParentCommentId = (*comment_dto.ParentCommentID).String()
	}

	return proto
}

type TagDTO struct {
	model.PostTags
}

type PostDTO struct {
	model.SubredditPosts
	Comments []CommentDTO
	Tags     []model.PostTags
}

func (post_dto *PostDTO) mapToProto() *pb.Post {
	proto := &pb.Post{
		Id:              post_dto.ID.String(),
		Title:           post_dto.Title,
		Content:         post_dto.Content,
		CreatedAt:       timestamppb.New(post_dto.CreatedAt),
		NumberOfVotes:   post_dto.NumberOfVotes,
		IsPinned:        post_dto.IsPinned,
		OwnerHandle:     post_dto.OwnerHandle,
		SubredditHandle: post_dto.SubredditHandle,
	}

	if post_dto.Image == nil {
		proto.Image = []byte("")
	} else {
		proto.Image = *post_dto.Image
	}

	for _, comment := range post_dto.Comments {
		proto.Comments = append(proto.Comments, comment.mapToProto())
	}

	for _, tag := range post_dto.Tags {
		proto.Tags = append(proto.Tags, tag.TagName)
	}
	return proto
}

type SubredditDTO struct {
	model.Subreddits
	Posts       []PostDTO
	JoinedUsers []model.SubredditUsers
}

func (sub_dto *SubredditDTO) mapToProto() *pb.Subreddit {
	proto := &pb.Subreddit{
		Handle:    sub_dto.Handle,
		Title:     sub_dto.Title,
		About:     sub_dto.About,
		Avatar:    sub_dto.Avatar,
		Rules:     sub_dto.Rules,
		CreatedAt: timestamppb.New(sub_dto.CreatedAt),
	}

	for _, post := range sub_dto.Posts {
		proto.Posts = append(proto.Posts, post.mapToProto())
	}

	for _, user := range sub_dto.JoinedUsers {
		if user.IsAdmin {
			proto.AdminsHandles = append(proto.AdminsHandles, user.UserHandle)
		} else {
			proto.UsersHandles = append(proto.UsersHandles, user.UserHandle)
		}
	}

	return proto
}
