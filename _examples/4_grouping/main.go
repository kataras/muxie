package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/*path", func(w http.ResponseWriter, r *http.Request) {
		path := muxie.GetParam(w, "path")
		fmt.Fprintf(w, "Site Custom 404 Error Message\nPage of: '%s' was unable to be found", path)
	})

	// `Of` will return a child router which will have the "/profile"
	// as its prefix for its routes registered by `Handle/HandleFunc`,
	// a child can have a child as well, i.e
	// profileRouter := mux.Of("/profile")
	// [...]
	// profileLikesRouter := profileRouter.Of("/likes")
	// will have its prefix as: "/profile/likes"
	profileRouter := mux.Of("/profile")

	// request: http://localhost:8080/profile
	// response: "Profile Index"
	profileRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Profile Index")
	})

	// request: http://localhost:8080/profile/kataras
	// response: "Profile of username: 'kataras'"
	profileRouter.HandleFunc("/:username", func(w http.ResponseWriter, r *http.Request) {
		username := muxie.GetParam(w, "username")
		fmt.Fprintf(w, "Profile of username: '%s'", username)
	})

	// matches everyhing if nothing else found after the /profile or /profile/ (if PathCorrection is true),
	// so you can use it for custom 404 profile pages!
	// For example:
	// request: http://localhost:8080/profile/kataras/what
	// response:
	// Profile Page Custom 404 Error Message
	// Profile Page of: '/kataras/what' was unable to be found
	profileRouter.HandleFunc("/*path", func(w http.ResponseWriter, r *http.Request) {
		path := muxie.GetParam(w, "path")
		fmt.Fprintf(w, "Profile Page Custom 404 Error Message\nProfile Page of: '%s' was unable to be found", path)
	})

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", mux)
}
