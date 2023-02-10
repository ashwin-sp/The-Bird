package storage

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/social/storage/memory"
)

func CreatePost(username string, message string) (memory.Post, int) {
	createdPost := memory.AddPost(username, message)
	return createdPost, http.StatusOK
}

func DeletePost(username string, postID uuid.UUID) int {
	// fill create Post API logic
	memory.RemovePost(username, postID)
	return http.StatusOK
}

func UpdateFollowStatus(me_user string, follower_user string, isAdd bool) int {
	if me_user == follower_user {
		return http.StatusNotAcceptable
	} else {
		memory.UpdateUserFollowMap(me_user, follower_user, isAdd)
		return http.StatusOK
	}

}

func ViewCreatedPosts(username string, frompage int) ([]memory.Post, int) {
	results := memory.GetPostsFromUserPosts(username, frompage)
	return results, http.StatusOK
}

func ViewPersonalFeed(username string, frompage int) ([]memory.Post, int) {
	results := memory.GetPostsFromUserFeed(username, frompage)
	return results, http.StatusOK
}

func ClearMemory() {
	memory.ClearMemory()
}

func InitTestMode() {
	memory.InitTestMode()
}
