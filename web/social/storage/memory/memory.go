package memory

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)


var mutex = &sync.Mutex{}
var testMap = make(map[string]string)
var isTestMode = false


func ClearMemory() {
	testMap = make(map[string]string)
	mutex = &sync.Mutex{}
	isTestMode = false
}

func InitTestMode() {
	isTestMode = true
}

func UpdateUserFollowMap(me_user string, follower_user string, isAdd bool) {
	mutex.Lock()
	followersMap := mapJSONToFollowers(getDataRaft("userFollowersMap_" + me_user))
	followingMap := mapJSONToFollowers(getDataRaft("userFollowingMap_" + follower_user))
	if isAdd {

		if !followersMap[follower_user] {
			followersMap[follower_user] = true
			sendDataRaft("userFollowersMap_"+me_user, mapFollowersToJSON(followersMap))

			followingMap[me_user] = true
			sendDataRaft("userFollowingMap_"+follower_user, mapFollowersToJSON(followingMap))

			if isTestMode { // Testing concurrency requires a very long wait, hence hack for test
				feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + follower_user))
				createdPosts := mapJSONToPosts(getDataRaft("userPostsMap_" + me_user))

				feedPosts = append(feedPosts, createdPosts...)

				// userFeedMap[follower_user] = append(userFeedMap[follower_user], userPostsMap[me_user]...)
				sort.Slice(feedPosts, func(i, j int) bool {
					return feedPosts[i].Timestamp.After(feedPosts[j].Timestamp)
				})
				removeDuplicatePost(feedPosts)
				sendDataRaft("userFeedMap_"+follower_user, mapPostsToJSON(feedPosts))
			} else {
				// publish changes
				go func(follower_user string, me_user string) {
					mutex.Lock()
					feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + follower_user))
					createdPosts := mapJSONToPosts(getDataRaft("userPostsMap_" + me_user))

					feedPosts = append(feedPosts, createdPosts...)

					// userFeedMap[follower_user] = append(userFeedMap[follower_user], userPostsMap[me_user]...)
					sort.Slice(feedPosts, func(i, j int) bool {
						return feedPosts[i].Timestamp.After(feedPosts[j].Timestamp)
					})
					removeDuplicatePost(feedPosts)
					sendDataRaft("userFeedMap_"+follower_user, mapPostsToJSON(feedPosts))
					mutex.Unlock()
				}(follower_user, me_user)
			}

		}

	} else {
		delete(followersMap, follower_user)
		delete(followingMap, me_user)
		if isTestMode {
			deletePostFromUserFeedByUserName(follower_user, me_user)
		} else {
			go func(follower_user string, me_user string) {
				mutex.Lock()
				deletePostFromUserFeedByUserName(follower_user, me_user)
				mutex.Unlock()
			}(follower_user, me_user)
		}
	}
	mutex.Unlock()
}

func AddPost(username string, message string) Post {
	mutex.Lock()
	defer mutex.Unlock()

	postToBeInserted := Post{Username: username, Message: message, PostID: uuid.New(), Timestamp: time.Now()}

	posts := mapJSONToPosts(getDataRaft("userPostsMap_" + username))
	// fmt.Printf(" Posts before %v\n", posts)
	posts = prependPost(posts, postToBeInserted)
	// fmt.Printf(" Posts after %v\n", posts)
	sendDataRaft("userPostsMap_"+username, mapPostsToJSON(posts))

	followersMap := mapJSONToFollowers(getDataRaft("userFollowersMap_" + username))
	for followerUserName := range followersMap {
		if isTestMode {
			feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + followerUserName))
			feedPosts = append(feedPosts, postToBeInserted)
			sort.Slice(feedPosts, func(i, j int) bool {
				return feedPosts[i].Timestamp.After(feedPosts[j].Timestamp)
			})
			removeDuplicatePost(feedPosts)
			sendDataRaft("userFeedMap_"+followerUserName, mapPostsToJSON(feedPosts))
		} else {
			go func(followerUserName string, postToBeInserted Post) {
				mutex.Lock()
				feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + followerUserName))
				feedPosts = append(feedPosts, postToBeInserted)
				sort.Slice(feedPosts, func(i, j int) bool {
					return feedPosts[i].Timestamp.After(feedPosts[j].Timestamp)
				})
				removeDuplicatePost(feedPosts)
				sendDataRaft("userFeedMap_"+followerUserName, mapPostsToJSON(feedPosts))
				mutex.Unlock()
			}(followerUserName, postToBeInserted)
		}

	}
	return posts[0] // just to ensure it is inserted properly
}

func RemovePost(username string, postID uuid.UUID) {
	mutex.Lock()
	i := 0 // output index

	posts := mapJSONToPosts(getDataRaft("userPostsMap_" + username))

	for _, post := range posts {
		if post.PostID != postID {
			// copy and increment index
			posts[i] = post
			i++
		}
	}
	posts = posts[:i]
	sendDataRaft("userPostsMap_"+username, mapPostsToJSON(posts))

	followersMap := mapJSONToFollowers(getDataRaft("userFollowersMap_" + username))

	// Publish changes - might need a background defer! - for now heavy suboptimal solution, will replace it with better one in further deliverable
	for followerUserName := range followersMap {
		if isTestMode {
			deletePostFromUserFeedByPostID(followerUserName, postID)
		} else {
			go func(followerUserName string, postID uuid.UUID) {
				mutex.Lock()
				deletePostFromUserFeedByPostID(followerUserName, postID)
				mutex.Unlock()
			}(followerUserName, postID)
		}
	}
	mutex.Unlock()
}

func deletePostFromUserFeedByPostID(username string, postID uuid.UUID) {
	// mutex.Lock()
	// defer mutex.Unlock()
	i := 0 // output index
	feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + username))
	for _, post := range feedPosts {
		if post.PostID != postID {
			// copy and increment index
			feedPosts[i] = post
			i++
		}
	}
	feedPosts = feedPosts[:i]
	sendDataRaft("userFeedMap_"+username, mapPostsToJSON(feedPosts))
}

func deletePostFromUserFeedByUserName(username string, unfollowed string) {
	// mutex.Lock()
	// defer mutex.Unlock()
	i := 0 // output index
	feedPosts := mapJSONToPosts(getDataRaft("userFeedMap_" + username))
	for _, post := range feedPosts {
		if post.Username != unfollowed {
			// copy and increment index
			feedPosts[i] = post
			i++
		}
	}
	feedPosts = feedPosts[:i]
	sendDataRaft("userFeedMap_"+username, mapPostsToJSON(feedPosts))
}

func GetPostsFromUserPosts(username string, frompage int) []Post {
	mutex.Lock()
	defer mutex.Unlock()
	postList := mapJSONToPosts(getDataRaft("userPostsMap_" + username))
	if frompage < len(postList) {
		var topage = frompage + 20
		if topage > len(postList) {
			topage = len(postList)
		}

		postList = postList[frompage:topage]
		return postList
	} else {
		return make([]Post, 0)
	}
}

func GetPostsFromUserFeed(username string, frompage int) []Post {
	mutex.Lock()
	defer mutex.Unlock()
	postList := mapJSONToPosts(getDataRaft("userFeedMap_" + username))
	// postList = removeDuplicatePost(postList)
	if frompage < len(postList) {
		var topage = frompage + 20
		if topage > len(postList) {
			topage = len(postList)
		}
		postList = postList[frompage:topage]
		return postList
	} else {
		return make([]Post, 0)
	}
}

// Reverse returns the reverse order for data.
func reverse(s []Post) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func removeDuplicatePost(postSlice []Post) []Post {
	allKeys := make(map[Post]bool)
	list := []Post{}
	for _, item := range postSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func prependPost(x []Post, y Post) []Post {
	x = append(x, Post{})
	copy(x[1:], x)
	x[0] = y
	return x
}


func mapPostsToJSON(posts []Post) string {
	jsonString, err := json.Marshal(posts)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

func mapFollowersToJSON(mapObj map[string]bool) string {
	jsonString, err := json.Marshal(mapObj)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

func mapJSONToPosts(jsonStr string) []Post {
	posts := []Post{}
	bytesJson := []byte(jsonStr)
	json.Unmarshal(bytesJson, &posts)
	return posts
}

func mapJSONToFollowers(jsonStr string) map[string]bool {
	mapObj := make(map[string]bool)
	bytesJson := []byte(jsonStr)
	json.Unmarshal(bytesJson, &mapObj)
	return mapObj
}

func sendDataRaft(key string, value string) {
	// mutex.Lock()
	if isTestMode {
		testMap[key] = value
	} else {
		url := "http://127.0.0.1:12380/" + key
		cmd := exec.Command("curl", "-L", url, "-XPUT", "-d", value)
		cmd.Run()
	}
	// mutex.Unlock()
}

func getDataRaft(key string) string {
	// mutex.Lock()
	// defer mutex.Unlock()
	if isTestMode {
		return testMap[key]
	} else {
		url := "http://127.0.0.1:12380/" + key
		req, _ := http.NewRequest("GET", url, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return testMap[key]
		}
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return string(body)
	}
}
