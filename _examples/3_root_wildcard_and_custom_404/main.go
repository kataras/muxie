package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	// matches everyhing if nothing else found, so you can use it for custom 404 main pages!
	mux.HandleFunc("/*path", func(w http.ResponseWriter, r *http.Request) {
		path := muxie.GetParam(w, "path")
		fmt.Fprintf(w, "Site Custom 404 Error Message\nPage of: '%s' was unable to be found", path)
	})
	mux.HandleFunc("/", indexHandler)

	// request: http://localhost:8080/profile
	// response: "Profile Index"
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Profile Index")
	})

	// request: http://localhost:8080/profile/kataras
	// response: "Profile of username: 'kataras'"
	mux.HandleFunc("/profile/:username", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/profile/*path", func(w http.ResponseWriter, r *http.Request) {
		path := muxie.GetParam(w, "path")
		fmt.Fprintf(w, "Profile Page Custom 404 Error Message\nProfile Page of: '%s' was unable to be found", path)
	})

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf8")
	fmt.Fprintf(w, "This is the <strong>%s</strong>", "index page")
}
