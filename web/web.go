package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/auth"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/social"
	"google.golang.org/grpc"
)

// Test to just see start of server!
func SayhelloName(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Please implement appropriate API or check your API endpoint!") // send data to client side
}

func SignUpAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+AUTH_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := auth.NewAuthServiceClient(conn)

	decoder := json.NewDecoder(r.Body)
	var credential auth.Credentials
	jsonerr := decoder.Decode(&credential)
	if jsonerr != nil {
		panic(jsonerr)
	}

	username, password := credential.Username, credential.Password

	token, resperr := c.SignUp(context.Background(), &auth.AuthMsg{Username: username, Password: password})
	if resperr != nil {
		log.Fatalf("Error when calling SayHello: %s", resperr)
	}

	if token.RefreshToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  "refreshtoken",
			Value: token.RefreshToken,
		})
	}
	if token.AccessToken != "" {
		expirationTime := time.Now().Add(EXPIRY_TIME) // Double door!
		http.SetCookie(w, &http.Cookie{
			Name:    "accesstoken",
			Value:   token.AccessToken,
			Expires: expirationTime,
		})
	}
	w.WriteHeader(int(token.Status))
	json.NewEncoder(w).Encode(token)
}

func LoginAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+AUTH_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := auth.NewAuthServiceClient(conn)

	decoder := json.NewDecoder(r.Body)
	var credential auth.Credentials
	err = decoder.Decode(&credential)
	if err != nil {
		panic(err)
	}

	username, password := credential.Username, credential.Password

	token, resperr := c.Login(context.Background(), &auth.AuthMsg{Username: username, Password: password})
	if resperr != nil {
		log.Fatalf("Error when calling SayHello: %s", resperr)
	}

	if token.RefreshToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  "refreshtoken",
			Value: token.RefreshToken,
		})
	}
	if token.AccessToken != "" {
		expirationTime := time.Now().Add(EXPIRY_TIME) // Double door!
		http.SetCookie(w, &http.Cookie{
			Name:    "accesstoken",
			Value:   token.AccessToken,
			Expires: expirationTime,
		})
	}
	w.WriteHeader(int(token.Status))
	json.NewEncoder(w).Encode(token)

}

func MaintainSessionAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+AUTH_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := auth.NewAuthServiceClient(conn)

	accesstoken := ""
	cookie, err := r.Cookie("accesstoken")
	if err == nil {
		accesstoken = cookie.Value
	} else {
		if err != http.ErrNoCookie {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	cookie, err = r.Cookie("refreshtoken")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	refreshtoken := cookie.Value

	token, resperr := c.MaintainSession(context.Background(), &auth.AuthCookieMsg{RefreshToken: refreshtoken, AccessToken: accesstoken})
	if resperr != nil {
		log.Fatalf("Error when calling SayHello: %s", resperr)
	}

	if token.AccessToken != "" {
		expirationTime := time.Now().Add(EXPIRY_TIME) // Double door!
		http.SetCookie(w, &http.Cookie{
			Name:    "accesstoken",
			Value:   token.AccessToken,
			Expires: expirationTime,
		})
	}
	w.WriteHeader(int(token.Status))
	json.NewEncoder(w).Encode(token)
}

func LogoutAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+AUTH_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := auth.NewAuthServiceClient(conn)
	cookie, err := r.Cookie("refreshtoken")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	refreshtoken := cookie.Value

	token, resperr := c.Logout(context.Background(), &auth.Token{TokenData: refreshtoken})
	if resperr != nil {
		log.Fatalf("Error when calling SayHello: %s", resperr)
	}

	cookie, err = r.Cookie("accesstoken")
	if err == nil {
		accesstoken := cookie.Value
		token.AccessToken = accesstoken
	}

	if token.Status == http.StatusOK {
		if token.RefreshToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:    "refreshtoken",
				Value:   token.RefreshToken,
				Expires: time.Now(),
			})
		}
		if token.AccessToken != "" {
			http.SetCookie(w, &http.Cookie{
				Name:    "accesstoken",
				Value:   token.AccessToken,
				Expires: time.Now(),
			})
		}
	}
	w.WriteHeader(int(token.Status))
}

func CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+SOCIAL_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := social.NewSocialServiceClient(conn)

	decoder := json.NewDecoder(r.Body)
	var postDetail social.PostResponse
	err = decoder.Decode(&postDetail)
	if err != nil {
		panic(err)
	}

	postMsg, resperr := c.CreatePost(context.Background(), &social.PostMsg{Username: postDetail.Username, Message: postDetail.Message})
	if resperr != nil {
		log.Fatalf("Error when calling CreatePostAPI: %s", resperr)
		return
	}
	// postMsg.Timestamp = postMsg.Timestamp.AsTime()
	uuid, error := uuid.Parse(postMsg.PostID)
	if error != nil {
		log.Fatalf("Error when converting to uuid CreatePostAPI: %s", resperr)
		return
	}
	postResponse := social.PostResponse{Username: postMsg.Username, Message: postMsg.Message, Timestamp: postMsg.Timestamp.AsTime(), PostID: uuid}
	w.WriteHeader(int(postMsg.Status))
	json.NewEncoder(w).Encode(postResponse)
}

func DeletePostAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+SOCIAL_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := social.NewSocialServiceClient(conn)
	decoder := json.NewDecoder(r.Body)
	var postDetail social.PostResponse
	err = decoder.Decode(&postDetail)
	if err != nil {
		panic(err)
	}
	status, resperr := c.DeletePost(context.Background(), &social.PostMsg{Username: postDetail.Username, PostID: postDetail.PostID.String()})
	if resperr != nil {
		log.Fatalf("Error when calling DeletePostAPI: %s", resperr)
	}
	w.WriteHeader(int(status.Data))
	// json.NewEncoder(w).Encode(postMsg)
}

func UpdateFollowStatusAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+SOCIAL_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := social.NewSocialServiceClient(conn)
	decoder := json.NewDecoder(r.Body)
	var followMap social.FollowMap
	err = decoder.Decode(&followMap)
	if err != nil {
		panic(err)
	}
	status, resperr := c.UpdateFollowStatus(context.Background(), &social.FollowMapMsg{Username: followMap.Username, Follower: followMap.Follower, Status: followMap.IsFollowing})
	if resperr != nil {
		log.Fatalf("Error when calling UpdateFollowStatus: %s", resperr)
	}
	w.WriteHeader(int(status.Data))
}

func ViewCreatedPostsAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+SOCIAL_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := social.NewSocialServiceClient(conn)
	decoder := json.NewDecoder(r.Body)
	var feedRequest social.FeedRequest
	err = decoder.Decode(&feedRequest)
	if err != nil {
		panic(err)
	}
	listofposts, resperr := c.ViewCreatedPosts(context.Background(), &social.FeedRequestMsg{Username: feedRequest.Username, FromPage: int32(feedRequest.FromPage)})
	// results, status := storage.ViewCreatedPosts(feedRequest.Username, feedRequest.FromPage)
	if resperr != nil {
		log.Fatalf("Error when calling ViewCreatedPostsAPI: %s", resperr)
	}
	//postResponse := social.PostResponse{Username: postMsg.Username, Message: postMsg.Message, Timestamp: postMsg.Timestamp.AsTime(), PostID: uuid}

	var postsResult []social.PostResponse
	for _, post := range listofposts.Value {
		uuid, error := uuid.Parse(post.PostID)
		if error != nil {
			log.Fatalf("Error when converting to uuid CreatePostAPI: %s", resperr)
			return
		}
		postsResult = append(postsResult, social.PostResponse{Username: post.Username, Message: post.Message, Timestamp: post.Timestamp.AsTime(), PostID: uuid})
	}
	w.WriteHeader(int(listofposts.Status))
	json.NewEncoder(w).Encode(postsResult)
}

func ViewPersonalFeedAPI(w http.ResponseWriter, r *http.Request) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":"+SOCIAL_PORT, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := social.NewSocialServiceClient(conn)
	decoder := json.NewDecoder(r.Body)
	var feedRequest social.FeedRequest
	err = decoder.Decode(&feedRequest)
	if err != nil {
		panic(err)
	}
	listofposts, resperr := c.ViewPersonalFeed(context.Background(), &social.FeedRequestMsg{Username: feedRequest.Username, FromPage: int32(feedRequest.FromPage)})
	// results, status := storage.ViewCreatedPosts(feedRequest.Username, feedRequest.FromPage)
	if resperr != nil {
		log.Fatalf("Error when calling ViewCreatedPostsAPI: %s", resperr)
	}
	//postResponse := social.PostResponse{Username: postMsg.Username, Message: postMsg.Message, Timestamp: postMsg.Timestamp.AsTime(), PostID: uuid}

	var postsResult []social.PostResponse
	for _, post := range listofposts.Value {
		uuid, error := uuid.Parse(post.PostID)
		if error != nil {
			log.Fatalf("Error when converting to uuid CreatePostAPI: %s", resperr)
			return
		}
		postsResult = append(postsResult, social.PostResponse{Username: post.Username, Message: post.Message, Timestamp: post.Timestamp.AsTime(), PostID: uuid})
	}
	w.WriteHeader(int(listofposts.Status))
	json.NewEncoder(w).Encode(postsResult)
}
