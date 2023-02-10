package social

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/social/storage"
)

func TestSelfPostsBasic(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8000/createpost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			post, status := storage.CreatePost(postDetail.Username, postDetail.Message)
			resp, err := httpmock.NewJsonResponse(status, post)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/viewcreatedposts",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var feedRequest FeedRequest
			err := decoder.Decode(&feedRequest)
			if err != nil {
				panic(err)
			}
			results, status := storage.ViewCreatedPosts(feedRequest.Username, feedRequest.FromPage)
			resp, err := httpmock.NewJsonResponse(status, results)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/deletepost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			status := storage.DeletePost(postDetail.Username, postDetail.PostID)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	storage.InitTestMode()
	username := "user4"
	messages := MakeRange(1, 25)

	for _, message := range messages {

		// fmt.Printf("Time check %v", time.Now().Add(time.Duration(message)*time.Minute).Unix())
		postBody, _ := json.Marshal(map[string]interface{}{
			"username": username,
			"message":  strconv.Itoa(message),
			// "timestamp": time.Now().Add(time.Duration(message) * time.Minute).Unix(),
		})
		requestBody := bytes.NewBuffer(postBody)
		//Leverage Go's HTTP Post function to make request

		resp, err := http.Post("http://localhost:8000/createpost", "application/json", requestBody)
		//Handle Error
		if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
		defer resp.Body.Close()
		//Read the response body
		decoder := json.NewDecoder(resp.Body)
		var postResponse PostResponse
		err = decoder.Decode(&postResponse)
		if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
		if resp.Status != "200" {
			t.Errorf("Unexpected error code %s", resp.Status)
		}

		if postResponse.PostID == uuid.Nil || postResponse.Username == "" || (postResponse.Timestamp == time.Time{}) {
			t.Errorf("Some parts of the response were empty %s, %s, %v", postResponse.Username, postResponse.PostID.String(), postResponse.Timestamp)
		}
	}

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8000/createpost"] != 25 {
		t.Errorf("Missed %d create API calls", 25-info["POST http://localhost:8000/createpost"])
	}

	postBody, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 0,
	})
	requestBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err := http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder := json.NewDecoder(resp.Body)
	var posts []PostResponse
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != 20 {
		t.Errorf("Returned incorrect count %d, expected 20", len(posts))
	}

	for i, post := range posts {
		if strconv.Itoa(messages[len(posts)+4-i]) != post.Message {
			t.Errorf("Created self posts not in expected order, expected %s, got %s", strconv.Itoa(messages[len(posts)+4-i]), post.Message)
			break
		}
	}

	// testing pagination
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 20,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != 5 {
		t.Errorf("Returned incorrect count %d, expected 5", len(posts))
	}

	for i, post := range posts {
		if strconv.Itoa(messages[len(posts)-1-i]) != post.Message {
			t.Errorf("Created self posts not in expected order, expected %s, got %s", strconv.Itoa(messages[len(posts)-1-i]), post.Message)
			break
		}
	}

	// deleting post messages which are multiple of 3  from 1 to 20
	for _, post := range posts {
		result, err := strconv.Atoi(post.Message)
		if err == nil && result%3 == 0 {
			postBody, _ := json.Marshal(map[string]interface{}{
				"username": username,
				"postid":   post.PostID,
			})
			requestBody := bytes.NewBuffer(postBody)
			//Leverage Go's HTTP Post function to make request

			resp, err := http.Post("http://localhost:8000/deletepost", "application/json", requestBody)
			//Handle Error
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			defer resp.Body.Close()
			//Read the response body
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			if resp.Status != "200" {
				t.Errorf("Unexpected error code %s", resp.Status)
			}
		} else if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
	}

	// we will just check for timestamp order now
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 0,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != 20 {
		t.Errorf("Returned incorrect count %d, expected 20", len(posts))
	}

	currentposttimestamp, previousposttimestamp, status := IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
	}

	// var previousTime time.Time
	// for i, post := range posts {
	// 	if i == 0 {
	// 		previousTime = post.Timestamp
	// 		continue
	// 	}
	// 	if post.Timestamp.After(previousTime) {
	// 		t.Errorf("Timestamp not in order, currentposttimestamp %s, previousposttimestamp %s", post.Timestamp.String(), previousTime.String())
	// 		break
	// 	} else {
	// 		previousTime = post.Timestamp
	// 	}
	// }
	t.Cleanup(ClearMemory)
}

func TestSelfPostsConcurrent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "http://localhost:8000/createpost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			post, status := storage.CreatePost(postDetail.Username, postDetail.Message)
			resp, err := httpmock.NewJsonResponse(status, post)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/viewcreatedposts",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var feedRequest FeedRequest
			err := decoder.Decode(&feedRequest)
			if err != nil {
				panic(err)
			}
			results, status := storage.ViewCreatedPosts(feedRequest.Username, feedRequest.FromPage)
			resp, err := httpmock.NewJsonResponse(status, results)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/deletepost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			status := storage.DeletePost(postDetail.Username, postDetail.PostID)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)
	storage.InitTestMode()
	var wg sync.WaitGroup

	username := "user4"
	messages := MakeRange(1, 25)

	for _, message := range messages {
		wg.Add(1)
		go func(message int, username string) {
			postBody, _ := json.Marshal(map[string]interface{}{
				"username": username,
				"message":  strconv.Itoa(message),
				// "timestamp": time.Now().Add(time.Duration(message) * time.Minute).Unix(),
			})
			requestBody := bytes.NewBuffer(postBody)
			//Leverage Go's HTTP Post function to make request

			resp, err := http.Post("http://localhost:8000/createpost", "application/json", requestBody)
			//Handle Error
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			defer resp.Body.Close()
			//Read the response body
			decoder := json.NewDecoder(resp.Body)
			var postResponse PostResponse
			err = decoder.Decode(&postResponse)
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			if resp.Status != "200" {
				t.Errorf("Unexpected error code %s", resp.Status)
			}

			if postResponse.PostID == uuid.Nil || postResponse.Username == "" || (postResponse.Timestamp == time.Time{}) {
				t.Errorf("Some parts of the response were empty %s, %s, %v", postResponse.Username, postResponse.PostID.String(), postResponse.Timestamp)
			}
			wg.Done()
		}(message, username)
	}
	wg.Wait()

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8000/createpost"] != 25 {
		t.Errorf("Missed %d create API calls", 25-info["POST http://localhost:8000/createpost"])
	}

	postBody, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 0,
	})
	requestBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err := http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder := json.NewDecoder(resp.Body)
	var posts []PostResponse
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	// log.Printf("\n")
	// for _, post := range posts {
	// 	log.Printf("%s by %s", post.Message, post.Username)
	// }
	// log.Printf("\n")

	if len(posts) != 20 {
		t.Errorf("Returned incorrect count %d, expected 20", len(posts))
	}

	currentposttimestamp, previousposttimestamp, status := IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
	}

	// testing pagination
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 20,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	// log.Printf("\n")
	// for _, post := range posts {
	// 	log.Printf("%s by %s", post.Message, post.Username)
	// }
	// log.Printf("\n")

	if len(posts) != 5 {
		t.Errorf("Returned incorrect count %d, expected 5", len(posts))
	}

	currentposttimestamp, previousposttimestamp, status = IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
	}

	// deleting post messages which are multiple of 3  from 1 to 20
	for _, post := range posts {
		result, err := strconv.Atoi(post.Message)
		if err == nil && result%3 == 0 {
			wg.Add(1)
			go func(username string, post PostResponse) {
				postBody, _ := json.Marshal(map[string]interface{}{
					"username": username,
					"postid":   post.PostID,
				})
				requestBody := bytes.NewBuffer(postBody)
				//Leverage Go's HTTP Post function to make request

				resp, err := http.Post("http://localhost:8000/deletepost", "application/json", requestBody)
				//Handle Error
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				defer resp.Body.Close()
				//Read the response body
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				if resp.Status != "200" {
					t.Errorf("Unexpected error code %s", resp.Status)
				}
				wg.Done()
			}(username, post)
		} else if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
	}
	wg.Wait()

	// we will just check for timestamp order now
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": username,
		"frompage": 0,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewcreatedposts", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != 20 {
		t.Errorf("Returned incorrect count %d, expected 20", len(posts))
	}

	currentposttimestamp, previousposttimestamp, status = IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
	}

	// var previousTime time.Time
	// for i, post := range posts {
	// 	if i == 0 {
	// 		previousTime = post.Timestamp
	// 		continue
	// 	}
	// 	if post.Timestamp.After(previousTime) {
	// 		t.Errorf("Timestamp not in order, currentposttimestamp %s, previousposttimestamp %s", post.Timestamp.String(), previousTime.String())
	// 		break
	// 	} else {
	// 		previousTime = post.Timestamp
	// 	}
	// }
	t.Cleanup(ClearMemory)
}

func TestFeedBasic(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	storage.InitTestMode()

	httpmock.RegisterResponder("POST", "http://localhost:8000/updatefollow",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var followMap FollowMap
			err := decoder.Decode(&followMap)
			if err != nil {
				panic(err)
			}

			status := storage.UpdateFollowStatus(followMap.Username, followMap.Follower, followMap.IsFollowing)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/createpost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			post, status := storage.CreatePost(postDetail.Username, postDetail.Message)
			resp, err := httpmock.NewJsonResponse(status, post)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/deletepost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			status := storage.DeletePost(postDetail.Username, postDetail.PostID)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/viewpersonalfeeds",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var feedRequest FeedRequest
			err := decoder.Decode(&feedRequest)
			if err != nil {
				panic(err)
			}

			results, status := storage.ViewPersonalFeed(feedRequest.Username, feedRequest.FromPage)
			resp, err := httpmock.NewJsonResponse(status, results)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	otherusers := []string{"user1", "user2", "user3"}

	for _, otheruser := range otherusers {
		postBody, _ := json.Marshal(map[string]interface{}{
			"username":    otheruser,
			"follower":    "user4",
			"isfollowing": true,
		})
		requestBody := bytes.NewBuffer(postBody)
		//Leverage Go's HTTP Post function to make request

		resp, err := http.Post("http://localhost:8000/updatefollow", "application/json", requestBody)
		//Handle Error
		if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
		defer resp.Body.Close()

		if resp.Status != "200" {
			t.Errorf("Unexpected error code %s", resp.Status)
		}
	}

	messages := MakeRange(1, 5)
	for _, otheruser := range otherusers {
		for _, message := range messages {
			postBody, _ := json.Marshal(map[string]interface{}{
				"username": otheruser,
				"message":  strconv.Itoa(message),
				// "timestamp": time.Now().Add(time.Duration(message) * time.Minute).Unix(),
			})
			requestBody := bytes.NewBuffer(postBody)
			//Leverage Go's HTTP Post function to make request

			resp, err := http.Post("http://localhost:8000/createpost", "application/json", requestBody)
			//Handle Error
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			defer resp.Body.Close()
			//Read the response body
			decoder := json.NewDecoder(resp.Body)
			var postResponse PostResponse
			err = decoder.Decode(&postResponse)
			if err != nil {
				t.Errorf("An Error Occured %v", err)
			}
			if resp.Status != "200" {
				t.Errorf("Unexpected error code %s", resp.Status)
			}

			if postResponse.PostID == uuid.Nil || postResponse.Username == "" || (postResponse.Timestamp == time.Time{}) {
				t.Errorf("Some parts of the response were empty %s, %s, %v", postResponse.Username, postResponse.PostID.String(), postResponse.Timestamp)
			}
		}
	}

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8000/createpost"] != 15 {
		t.Errorf("Missed %d create API calls", 15-info["POST http://localhost:8000/createpost"])
	}

	// time.Sleep(5 * time.Second) // Just to wait for background publish to end

	// we will just check for timestamp order now
	postBody, _ := json.Marshal(map[string]interface{}{
		"username": "user4",
		"frompage": 0,
	})
	requestBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err := http.Post("http://localhost:8000/viewpersonalfeeds", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	var posts []PostResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != info["POST http://localhost:8000/createpost"] {
		t.Errorf("Returned incorrect count %d, expected %d", len(posts), info["POST http://localhost:8000/createpost"])
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v", post.Message, post.Username, post.PostID)
		}
		log.Printf("\n")
	}

	currentposttimestamp, previousposttimestamp, status := IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v time %v", post.Message, post.Username, post.PostID, post.Timestamp)
		}
		log.Printf("\n")
	}

	for _, otheruser := range otherusers {
		for _, post := range posts {
			if post.Message == strconv.Itoa(3) {
				postBody, _ := json.Marshal(map[string]interface{}{
					"username": otheruser,
					"postid":   post.PostID,
				})
				requestBody := bytes.NewBuffer(postBody)
				//Leverage Go's HTTP Post function to make request

				resp, err := http.Post("http://localhost:8000/deletepost", "application/json", requestBody)
				//Handle Error
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				defer resp.Body.Close()
				//Read the response body
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				if resp.Status != "200" {
					t.Errorf("Unexpected error code %s", resp.Status)
				}
			}
		}
	}

	// time.Sleep(5 * time.Second) // Just to wait for background publish to end
	// we will just check for timestamp order now
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": "user4",
		"frompage": 0,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewpersonalfeeds", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != info["POST http://localhost:8000/createpost"]-3 {
		t.Errorf("Returned incorrect count %d, expected %d", len(posts), info["POST http://localhost:8000/createpost"]-3)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v", post.Message, post.Username, post.PostID)
		}
		log.Printf("\n")
	}

	currentposttimestamp, previousposttimestamp, status = IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v time %v", post.Message, post.Username, post.PostID, post.Timestamp)
		}
		log.Printf("\n")
	}
	t.Cleanup(ClearMemory)
}

func TestFeedConcurrent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	storage.InitTestMode()

	httpmock.RegisterResponder("POST", "http://localhost:8000/updatefollow",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var followMap FollowMap
			err := decoder.Decode(&followMap)
			if err != nil {
				panic(err)
			}

			status := storage.UpdateFollowStatus(followMap.Username, followMap.Follower, followMap.IsFollowing)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/createpost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			post, status := storage.CreatePost(postDetail.Username, postDetail.Message)
			resp, err := httpmock.NewJsonResponse(status, post)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/deletepost",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var postDetail PostResponse
			err := decoder.Decode(&postDetail)
			if err != nil {
				panic(err)
			}
			status := storage.DeletePost(postDetail.Username, postDetail.PostID)
			resp, err := httpmock.NewJsonResponse(status, ``)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/viewpersonalfeeds",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var feedRequest FeedRequest
			err := decoder.Decode(&feedRequest)
			if err != nil {
				panic(err)
			}

			results, status := storage.ViewPersonalFeed(feedRequest.Username, feedRequest.FromPage)
			resp, err := httpmock.NewJsonResponse(status, results)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	otherusers := []string{"user1", "user2", "user3"}

	for _, otheruser := range otherusers {
		postBody, _ := json.Marshal(map[string]interface{}{
			"username":    otheruser,
			"follower":    "user4",
			"isfollowing": true,
		})
		requestBody := bytes.NewBuffer(postBody)
		//Leverage Go's HTTP Post function to make request

		resp, err := http.Post("http://localhost:8000/updatefollow", "application/json", requestBody)
		//Handle Error
		if err != nil {
			t.Errorf("An Error Occured %v", err)
		}
		defer resp.Body.Close()

		if resp.Status != "200" {
			t.Errorf("Unexpected error code %s", resp.Status)
		}
	}

	// time.Sleep(5 * time.Second)
	var wg sync.WaitGroup

	messages := MakeRange(1, 5)
	for _, otheruser := range otherusers {
		for _, message := range messages {
			wg.Add(1)
			go func(message int, otheruser string) {
				postBody, _ := json.Marshal(map[string]interface{}{
					"username": otheruser,
					"message":  strconv.Itoa(message),
					// "timestamp": time.Now().Add(time.Duration(message) * time.Minute).Unix(),
				})
				requestBody := bytes.NewBuffer(postBody)
				//Leverage Go's HTTP Post function to make request

				resp, err := http.Post("http://localhost:8000/createpost", "application/json", requestBody)
				//Handle Error
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				defer resp.Body.Close()
				//Read the response body
				decoder := json.NewDecoder(resp.Body)
				var postResponse PostResponse
				err = decoder.Decode(&postResponse)
				if err != nil {
					t.Errorf("An Error Occured %v", err)
				}
				if resp.Status != "200" {
					t.Errorf("Unexpected error code %s", resp.Status)
				}

				if postResponse.PostID == uuid.Nil || postResponse.Username == "" || (postResponse.Timestamp == time.Time{}) {
					t.Errorf("Some parts of the response were empty %s, %s, %v", postResponse.Username, postResponse.PostID.String(), postResponse.Timestamp)
				}
				wg.Done()
			}(message, otheruser)
		}
	}

	wg.Wait()

	// time.Sleep(5 * time.Second) // Just to wait for background publish to end

	info := httpmock.GetCallCountInfo()
	if info["POST http://localhost:8000/createpost"] != 15 {
		t.Errorf("Missed %d create API calls", 15-info["POST http://localhost:8000/createpost"])
	}

	// we will just check for timestamp order now
	postBody, _ := json.Marshal(map[string]interface{}{
		"username": "user4",
		"frompage": 0,
	})
	requestBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err := http.Post("http://localhost:8000/viewpersonalfeeds", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	var posts []PostResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != info["POST http://localhost:8000/createpost"] {

		t.Errorf("Returned incorrect count %d, expected %d", len(posts), info["POST http://localhost:8000/createpost"])
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v", post.Message, post.Username, post.PostID)
		}
		log.Printf("\n")
	}

	currentposttimestamp, previousposttimestamp, status := IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v time %v", post.Message, post.Username, post.PostID, post.Timestamp)
		}
		log.Printf("\n")
	}

	for _, otheruser := range otherusers {
		for _, post := range posts {
			if post.Message == strconv.Itoa(3) {
				wg.Add(1)
				go func(otheruser string, post PostResponse) {
					postBody, _ := json.Marshal(map[string]interface{}{
						"username": otheruser,
						"postid":   post.PostID,
					})
					requestBody := bytes.NewBuffer(postBody)
					//Leverage Go's HTTP Post function to make request

					resp, err := http.Post("http://localhost:8000/deletepost", "application/json", requestBody)
					//Handle Error
					if err != nil {
						t.Errorf("An Error Occured %v", err)
					}
					defer resp.Body.Close()
					//Read the response body
					if err != nil {
						t.Errorf("An Error Occured %v", err)
					}
					if resp.Status != "200" {
						t.Errorf("Unexpected error code %s", resp.Status)
					}
					wg.Done()
				}(otheruser, post)
			}
		}
	}
	wg.Wait()

	// time.Sleep(5 * time.Second) // Just to wait for background publish to end

	// we will just check for timestamp order now
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": "user4",
		"frompage": 0,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/viewpersonalfeeds", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}

	if len(posts) != info["POST http://localhost:8000/createpost"]-3 {
		t.Errorf("Returned incorrect count %d, expected %d", len(posts), info["POST http://localhost:8000/createpost"]-3)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v time %v", post.Message, post.Username, post.PostID, post.Timestamp)
		}
		log.Printf("\n")
	}

	currentposttimestamp, previousposttimestamp, status = IsInLatestOrder(posts)
	if !status {
		t.Errorf("Timestamp not in order, currentposttimestamp %v, previousposttimestamp %v", currentposttimestamp, previousposttimestamp)
		log.Printf("\n")
		for _, post := range posts {
			log.Printf("%s by %s with id %v time %v", post.Message, post.Username, post.PostID, post.Timestamp)
		}
		log.Printf("\n")
	}
	t.Cleanup(ClearMemory)
}
