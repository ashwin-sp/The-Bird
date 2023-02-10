package memory

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Can add more properties on the fly!
type User struct {
	Passhash string
}

type OAuth struct {
	RefreshToken string
}

type Session struct {
	AccessToken string
}
type Claims struct {
	Username     string
	RefreshToken string
	jwt.RegisteredClaims
}

var jwtKey = []byte(SECRET_KEY)

var testMap = make(map[string]string)

// var userMap = make(map[string]User)
var mutex = &sync.Mutex{}
var isTestMode = true

// var oauthMap = make(map[string][]OAuth)

// reverse index map style logic refresh and access have 1-1 relation so its alright!
// var sessionMap = make(map[string]Session)
// var reverseSessionMap = make(map[string]OAuth)

func GetUserObject(username string) User {
	mutex.Lock()
	defer mutex.Unlock()
	return mapJSONToUser(getDataRaft("userMap_" + username))
}

func SetUserObject(username string, password string) User {
	mutex.Lock()
	defer mutex.Unlock()

	passhash, err := bcrypt.GenerateFromPassword([]byte(password), HASH_COST)
	if err != nil {
		return User{}
	}
	user := User{Passhash: string(passhash)}
	sendDataRaft("userMap_"+username, mapUserToJSON(user))
	return user
}

func GenerateRefreshToken(username string) string {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return ""
	}

	return tokenString
}

func GenerateAccessToken(refreshtoken string) string {
	expirationTime := time.Now().Add(EXPIRY_TIME) // 1 min expiry just for check
	claims := &Claims{
		RefreshToken: refreshtoken,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return ""
	}

	return tokenString
}

func CheckPassword(passhash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passhash), []byte(password))
	return err == nil
}

func GeneratePasshash(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}

func CheckAccessToken(token string) int {
	// indirect validation of an accesstoken presence

	if (mapJSONToOAuth(getDataRaft("reverseSessionMap_"+token)) != OAuth{}) {
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				// log.Fatal("Signature Invalid\n")
				return http.StatusUnauthorized
			}

			// log.Fatal("Expired\n")
			return http.StatusUnauthorized
		}
		if !tkn.Valid {
			// log.Fatal("Not valid\n")
			return http.StatusUnauthorized
		}
	} else {
		return http.StatusNotFound
	}
	// if claims.ExpiresAt.Time.Before(time.Now()) {
	// 	log.Fatal("Expired \n")
	// 	return http.StatusBadRequest
	// }
	return http.StatusOK
}

func CreateNewRefreshToken(username string) (string, int) {
	mutex.Lock()
	defer mutex.Unlock()

	if (mapJSONToUser(getDataRaft("userMap_"+username)) != User{}) {
		refreshtoken := GenerateRefreshToken(username)
		if refreshtoken == "" {
			return "", http.StatusInternalServerError
		}
		sendDataRaft("sessionMap_"+refreshtoken, mapSessionToJSON(Session{AccessToken: "*"}))
		return refreshtoken, http.StatusOK
	} else {
		return "", http.StatusNotFound
	}
}

func CreateNewSession(refreshtoken string) (string, int) {
	mutex.Lock()
	defer mutex.Unlock()
	// indirect validation of a refreshtoken presence
	sessionObj := mapJSONToSession(getDataRaft("sessionMap_" + refreshtoken))
	if (sessionObj != Session{}) {
		accesstoken := GenerateAccessToken(refreshtoken)
		if accesstoken == "" {
			return "", http.StatusInternalServerError
		}
		sendDataRaft("reverseSessionMap_"+sessionObj.AccessToken, mapOAuthToJSON(OAuth{}))
		sendDataRaft("sessionMap_"+refreshtoken, mapSessionToJSON(Session{AccessToken: accesstoken}))
		sendDataRaft("reverseSessionMap_"+accesstoken, mapOAuthToJSON(OAuth{RefreshToken: refreshtoken}))
		return accesstoken, http.StatusOK
	} else {
		// logged out user
		return "", http.StatusNotFound
	}

}

func DeleteRefreshToken(refreshtoken string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	sessionObj := mapJSONToSession(getDataRaft("sessionMap_" + refreshtoken))

	if (sessionObj != Session{}) {
		sendDataRaft("sessionMap_"+refreshtoken, mapSessionToJSON(Session{}))
		sendDataRaft("reverseSessionMap_"+sessionObj.AccessToken, mapOAuthToJSON(OAuth{}))
		return true
	} else {
		return false
	}
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

func mapUserToJSON(user User) string {
	jsonString, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

func mapJSONToUser(jsonStr string) User {
	user := User{}
	bytesJson := []byte(jsonStr)
	json.Unmarshal(bytesJson, &user)
	return user
}

func mapJSONToSession(jsonStr string) Session {
	session := Session{}
	bytesJson := []byte(jsonStr)
	json.Unmarshal(bytesJson, &session)
	return session
}

func mapSessionToJSON(session Session) string {
	jsonString, err := json.Marshal(session)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

func mapOAuthToJSON(oauth OAuth) string {
	jsonString, err := json.Marshal(oauth)
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}

func mapJSONToOAuth(jsonStr string) OAuth {
	oauth := OAuth{}
	bytesJson := []byte(jsonStr)
	json.Unmarshal(bytesJson, &oauth)
	return oauth
}
