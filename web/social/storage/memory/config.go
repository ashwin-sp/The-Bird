package memory

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	PostID    uuid.UUID `json:"postid"`
	Username  string    `json:"username"`
	// add more data if needed
}

func GetTime(timestamp string) time.Time {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Unix(i, 0)
}
