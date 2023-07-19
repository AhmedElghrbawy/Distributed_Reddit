package main

import (
	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/model"
	"github.com/ahmedelghrbawy/replicated_db/pkg/jet_db/public/table"
	pb "github.com/ahmedelghrbawy/replicated_db/pkg/rdb_grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentDTO struct {
	model.SubredditComments
}

// since CommentDTO defines its embedded post as SubredditComments model,
// trying to use jet QRM won't work for UserComments model.
// the solution is to use table alias for UserComments table.
// anywhere you need to map a UserComments to a CommentDTO, you should use this alias
var AliasedUserComments = table.UserComments.AS("SubredditComments")

func (comment_dto *CommentDTO) mapToProto() *pb.Comment {
	proto := &pb.Comment{
		Id:            comment_dto.ID.String(),
		Content:       comment_dto.Content,
		NumberOfVotes: comment_dto.NumberOfVotes,
		OwnerHandle:   comment_dto.OwnerHandle,
		PostId:        comment_dto.PostID.String(),
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

// since PostDTO defines its embedded post as SubredditPosts model,
// trying to use jet QRM won't work for UserPosts model.
// the solution is to use table alias for UserPosts table.
// anywhere you need to map a UserPost to a PostDto, you should use this alias
var AliasedUserPosts = table.UserPosts.AS("SubredditPosts")

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

type UserDTO struct {
	model.Users
	Posts            []PostDTO
	Comments         []CommentDTO
	Followage        []model.Followage
	JoinedSubreddits []model.SubredditUsers
}

func (user_dto *UserDTO) mapToProto() *pb.User {
	proto := &pb.User{
		Handle:      user_dto.Handle,
		DisplayName: user_dto.DisplayName,
		Avatar:      user_dto.Avatar,
		CreatedAt:   timestamppb.New(user_dto.CreatedAt),
		Karma:       user_dto.Karma,
	}

	for _, post := range user_dto.Posts {
		proto.Posts = append(proto.Posts, post.mapToProto())
	}

	for _, comment := range user_dto.Comments {
		proto.Comments = append(proto.Comments, comment.mapToProto())
	}

	for _, followage := range user_dto.Followage {
		if followage.FollowerHandle == user_dto.Handle {
			proto.FollowingHandles = append(proto.FollowingHandles, followage.FollowedHandle)
		} else {
			proto.FollowedByHandles = append(proto.FollowedByHandles, followage.FollowerHandle)
		}
	}

	for _, sub := range user_dto.JoinedSubreddits {
		if sub.IsAdmin {
			proto.AdminedSubredditHandles = append(proto.AdminedSubredditHandles, sub.SubredditHandle)
		} else {
			proto.JoinedSubredditHandles = append(proto.JoinedSubredditHandles, sub.SubredditHandle)
		}
	}

	return proto
}
