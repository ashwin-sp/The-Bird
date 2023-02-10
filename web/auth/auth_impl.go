package auth

import (
	"net/http"

	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/auth/storage"
	"golang.org/x/net/context"
)

type Server struct {
	UnimplementedAuthServiceServer
}

func (s *Server) SignUp(ctx context.Context, in *AuthMsg) (*AuthMsgResponse, error) {
	username := in.Username
	password := in.Password
	var response = &AuthMsgResponse{AccessToken: "", RefreshToken: "", Status: int32(403)}
	token, status := storage.SignUp(username, password)
	response.AccessToken = token.AccessToken
	response.RefreshToken = token.RefreshToken
	response.Status = int32(status)
	return response, nil
}

func (s *Server) Login(ctx context.Context, in *AuthMsg) (*AuthMsgResponse, error) {
	username := in.Username
	password := in.Password
	var response = &AuthMsgResponse{AccessToken: "", RefreshToken: "", Status: int32(403)}
	token, status := storage.Login(username, password)
	response.AccessToken = token.AccessToken
	response.RefreshToken = token.RefreshToken
	response.Status = int32(status)
	return response, nil
}

func (s *Server) Logout(ctx context.Context, in *Token) (*AuthMsgResponse, error) {
	refreshToken := in.TokenData
	var response = &AuthMsgResponse{AccessToken: "", RefreshToken: "", Status: int32(403)}
	token, status := storage.Logout(refreshToken)
	response.AccessToken = token.AccessToken
	response.RefreshToken = token.RefreshToken
	response.Status = int32(status)
	return response, nil
}

func (s *Server) MaintainSession(ctx context.Context, c *AuthCookieMsg) (*AuthMsgResponse, error) {
	var response = &AuthMsgResponse{AccessToken: "", RefreshToken: "", Status: int32(403)}
	accesstoken := c.AccessToken
	status := storage.CheckSession(accesstoken)
	if status != http.StatusOK {
		refreshtoken := c.RefreshToken
		token, status := storage.CreateSession(refreshtoken)
		response.AccessToken = token.AccessToken
		response.RefreshToken = token.RefreshToken
		response.Status = int32(status)
	} else {
		response.Status = int32(status)
	}
	return response, nil
}
