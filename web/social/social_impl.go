package social

import (
	"github.com/google/uuid"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/social/storage"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	UnimplementedSocialServiceServer
}

func (s *Server) CreatePost(ctx context.Context, in *PostMsg) (*PostMsg, error) {
	message := in.Message
	username := in.Username
	data, status := storage.CreatePost(username, message)
	var response = &PostMsg{Timestamp: timestamppb.Now(), Message: data.Message, PostID: data.PostID.String(), Username: data.Username, Status: int32(status)}
	response.Status = int32(status)
	return response, nil
}

func (s *Server) DeletePost(ctx context.Context, in *PostMsg) (*Status, error) {
	PostID := in.PostID
	username := in.Username
	con_postID, _ := uuid.Parse(PostID)
	status := storage.DeletePost(username, con_postID)
	var response = &Status{Data: int32(status)}
	return response, nil
}

func (s *Server) UpdateFollowStatus(ctx context.Context, in *FollowMapMsg) (*Status, error) {
	status := storage.UpdateFollowStatus(in.Username, in.Follower, in.Status)
	var response = &Status{Data: int32(status)}
	return response, nil
}

func (s *Server) ViewCreatedPosts(ctx context.Context, in *FeedRequestMsg) (*ListOfPosts, error) {
	results, status := storage.ViewCreatedPosts(in.Username, int(in.FromPage))
	var response = &ListOfPosts{
		Value:                []*PostMsg{},
		Status:               int32(status),
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}
	for i := 0; i < len(results); i++ {
		tmp := PostMsg{
			Timestamp:            timestamppb.New(results[i].Timestamp),
			Message:              results[i].Message,
			PostID:               results[i].PostID.String(),
			Username:             results[i].Username,
			Status:               0,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		}
		response.Value = append(response.Value, &tmp)
	}
	response.Status = int32(status)
	return response, nil
}

func (s *Server) ViewPersonalFeed(ctx context.Context, in *FeedRequestMsg) (*ListOfPosts, error) {
	results, status := storage.ViewPersonalFeed(in.Username, int(in.FromPage))
	var response = &ListOfPosts{
		Value:                []*PostMsg{},
		Status:               int32(status),
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}
	for i := 0; i < len(results); i++ {
		tmp := PostMsg{
			Timestamp:            timestamppb.New(results[i].Timestamp),
			Message:              results[i].Message,
			PostID:               results[i].PostID.String(),
			Username:             results[i].Username,
			Status:               0,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		}
		response.Value = append(response.Value, &tmp)
	}
	response.Status = int32(status)
	return response, nil
}
