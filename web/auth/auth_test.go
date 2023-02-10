package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/auth/storage"
)

func TestUserLifeCycle(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// token := make(map[string]interface{}, 0)
	// Exact URL match
	httpmock.RegisterResponder("POST", "http://localhost:8000/signup",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var credential Credentials
			err := decoder.Decode(&credential)
			if err != nil {
				panic(err)
			}

			username, password := credential.Username, credential.Password
			token, status := storage.SignUp(username, password)
			resp, err := httpmock.NewJsonResponse(status, token)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/login",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var credential Credentials
			err := decoder.Decode(&credential)
			if err != nil {
				panic(err)
			}

			token, status := storage.Login(credential.Username, credential.Password)
			resp, err := httpmock.NewJsonResponse(status, token)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/maintainsession",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var token storage.Token
			err := decoder.Decode(&token)
			if err != nil {
				panic(err)
			}

			status := storage.CheckSession(token.AccessToken)
			if status != http.StatusOK {
				token, status := storage.CreateSession(token.RefreshToken)
				resp, err := httpmock.NewJsonResponse(status, token)
				if err != nil {
					return httpmock.NewStringResponse(500, ``), nil
				}
				return resp, nil
			}
			resp, err := httpmock.NewJsonResponse(status, storage.Token{})
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	httpmock.RegisterResponder("POST", "http://localhost:8000/logout",
		func(req *http.Request) (*http.Response, error) {
			decoder := json.NewDecoder(req.Body)
			var token storage.Token
			err := decoder.Decode(&token)
			if err != nil {
				panic(err)
			}

			token, status := storage.Logout(token.RefreshToken)
			resp, err := httpmock.NewJsonResponse(status, token)
			if err != nil {
				return httpmock.NewStringResponse(500, ``), nil
			}
			return resp, nil
		},
	)

	postBody, _ := json.Marshal(map[string]interface{}{
		"username": "user4",
		"password": "password4",
	})
	requestBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err := http.Post("http://localhost:8000/signup", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder := json.NewDecoder(resp.Body)
	var token storage.Token
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	refreshtoken := token.RefreshToken
	accesstoken := token.AccessToken
	if resp.Status != "200" {
		t.Errorf("Unexpected error code %s", resp.Status)
	}

	if refreshtoken == "" || accesstoken == "" {
		t.Errorf("Refresh/access token was empty %s, %s", refreshtoken, accesstoken)
	}

	postBody, _ = json.Marshal(map[string]interface{}{
		"username": "user4",
		"password": "password4",
	})
	requestBody = bytes.NewBuffer(postBody)
	resp, err = http.Post("http://localhost:8000/signup", "application/json", requestBody)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	if resp.Status != "403" { // user already exists so should return 403
		t.Errorf("User already exists so expected 403 but got %s", resp.Status)
	}

	postBody, _ = json.Marshal(map[string]interface{}{
		"username": "user4",
		"password": "password4",
	})
	requestBody = bytes.NewBuffer(postBody)
	resp, err = http.Post("http://localhost:8000/login", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	refreshtoken = token.RefreshToken
	accesstoken = token.AccessToken
	if resp.Status != "200" {
		t.Errorf("Unexpected error code %s", resp.Status)
	}

	if refreshtoken == "" || accesstoken == "" {
		t.Errorf("Refresh/access token was empty %s, %s", refreshtoken, accesstoken)
	}
	postBody, _ = json.Marshal(map[string]interface{}{
		"username": "user3",
		"password": "password4",
	})
	requestBody = bytes.NewBuffer(postBody)
	resp, err = http.Post("http://localhost:8000/login", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	if resp.Status != "404" { // Non existant user trying to login
		t.Errorf("User does not exist so expected 404 but got %s", resp.Status)
	}

	postBody, _ = json.Marshal(map[string]interface{}{
		"refreshtoken": refreshtoken,
		"accesstoken":  accesstoken,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/maintainsession", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	if resp.Status != "200" {
		t.Errorf("Unexpected error code %s", resp.Status)
	}
	if token.RefreshToken != "" || token.AccessToken != "" {
		t.Errorf("Should not have refreshed this early")
	}
	time.Sleep(12 * time.Second)

	postBody, _ = json.Marshal(map[string]interface{}{
		"refreshtoken": refreshtoken,
		"accesstoken":  accesstoken,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/maintainsession", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	if resp.Status != "200" {
		t.Errorf("Unexpected error code %s", resp.Status)
	}
	if token.RefreshToken == "" || token.AccessToken == "" {
		t.Errorf("Refresh failed")
	}

	postBody, _ = json.Marshal(map[string]interface{}{
		"refreshtoken": refreshtoken,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/logout", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	if resp.Status != "200" {
		t.Errorf("Unexpected error code %s", resp.Status)
	}
	if token.RefreshToken == "" {
		t.Errorf("Should have returned the deleted refresh token")
	}

	postBody, _ = json.Marshal(map[string]interface{}{
		"refreshtoken": refreshtoken,
	})
	requestBody = bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request

	resp, err = http.Post("http://localhost:8000/logout", "application/json", requestBody)
	//Handle Error
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		t.Errorf("An Error Occured %v", err)
	}
	if resp.Status != "404" { // refreshtoken should not be there
		t.Errorf("Unexpected error code %s", resp.Status)
	}
	if token.RefreshToken != "" {
		t.Errorf("Shouldn't have had the refreshtoken in memory")
	}

}
