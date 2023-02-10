package web

import (
	"log"
	"net/http"

	"github.com/rs/cors"
)

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", SayhelloName) // set router

	mux.HandleFunc("/signup", SignUpAPI) // signup api
	mux.HandleFunc("/login", LoginAPI)   // login api or if you had lost your refresh token
	mux.HandleFunc("/maintainsession", MaintainSessionAPI) // check + create
	mux.HandleFunc("/logout", LogoutAPI)                   // Delete refresh token itself
	mux.HandleFunc("/createpost", CreatePostAPI)
	mux.HandleFunc("/deletepost", DeletePostAPI)
	mux.HandleFunc("/updatefollow", UpdateFollowStatusAPI)
	mux.HandleFunc("/viewcreatedposts", ViewCreatedPostsAPI)
	mux.HandleFunc("/viewpersonalfeeds", ViewPersonalFeedAPI)

	handler := cors.Default().Handler(mux)
	err := http.ListenAndServe(":"+PORT, handler) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		log.Println("Listening on port: " + PORT)
	}
}
