package social

import (
	"time"

	"github.com/google/uuid"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/social/storage"
)

type PostResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	PostID    uuid.UUID `json:"postid"`
	Username  string    `json:"username"`
	// add more data if needed
}

type FeedRequest struct {
	Username string `json:"username"`
	FromPage int    `json:"frompage"`
	// can add timestamp support later
}

type FollowMap struct {
	Username    string `json:"username"`
	Follower    string `json:"follower"`
	IsFollowing bool   `json:"isfollowing"`
	// add more data if needed
}

func IsInLatestOrder(posts []PostResponse) (string, string, bool) {
	var previousTime time.Time
	for i, post := range posts {
		if i == 0 {
			previousTime = post.Timestamp
			continue
		}
		if post.Timestamp.After(previousTime) {
			return post.Timestamp.String(), previousTime.String(), false
		} else {
			previousTime = post.Timestamp
		}
	}
	return "", "", true
}

func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func ClearMemory() {
	storage.ClearMemory()
}
