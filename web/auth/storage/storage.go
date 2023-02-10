package storage

import (
	"net/http"

	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/auth/storage/memory"
)

const EXPIRY_TIME = memory.EXPIRY_TIME

type Token struct {
	RefreshToken string `json:"refreshtoken"`
	AccessToken  string `json:"accesstoken"`
}

func CheckSession(accesstoken string) int {
	return memory.CheckAccessToken(accesstoken)
}

func CreateSession(refreshtoken string) (Token, int) {
	token := Token{RefreshToken: refreshtoken}
	accesstoken, status := memory.CreateNewSession(refreshtoken)
	if status == http.StatusOK {
		token.AccessToken = accesstoken
	}
	return token, http.StatusOK
}

func SignUp(username string, password string) (Token, int) {
	user := memory.GetUserObject(username)
	if user.Passhash != "" {
		// User is already present, do login
		return Token{}, http.StatusForbidden
	} else {
		// Creating a new user!
		user := memory.SetUserObject(username, password)
		if user.Passhash != "" {
			refreshtoken, status := memory.CreateNewRefreshToken(username)
			if status != http.StatusOK {
				return Token{}, status
			} else {
				var token = Token{RefreshToken: refreshtoken}
				accesstoken, status := memory.CreateNewSession(refreshtoken)
				if status != http.StatusOK {
					return Token{}, status
				} else {
					token.AccessToken = accesstoken
					return token, status
				}
			}
		} else {
			return Token{}, http.StatusInternalServerError
		}
	}
}

func Login(username string, password string) (Token, int) {
	user := memory.GetUserObject(username)
	if user.Passhash != "" {
		// User is already present, do login
		if memory.CheckPassword(user.Passhash, password) {
			refreshtoken, status := memory.CreateNewRefreshToken(username)
			if status != http.StatusOK {
				return Token{}, status
			} else {
				var token = Token{RefreshToken: refreshtoken}
				accesstoken, status := memory.CreateNewSession(refreshtoken)
				if status != http.StatusOK {
					return Token{}, status
				} else {
					token.AccessToken = accesstoken
					return token, status
				}
			}
		} else {
			return Token{}, http.StatusUnauthorized
		}
	} else {
		// User not found!
		return Token{}, http.StatusNotFound
	}
}

func Logout(refreshtoken string) (Token, int) {
	if memory.DeleteRefreshToken(refreshtoken) {
		token := Token{RefreshToken: refreshtoken}
		return token, http.StatusOK
	} else {
		return Token{}, http.StatusNotFound
	}
}
