package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	// static root, matches http://localhost:8080
	// or http://localhost:8080/ (even if PathCorrection is false).
	mux.HandleFunc("/", indexHandler)

	// named parameter, matches /profile/$something_here
	// but NOT /profile/anything/here neither /profile
	// and /profile/ (if PathCorrection is true).
	mux.HandleFunc("/profile/:name", profileHandler)

	// named parameter followed by static segmet, matches /profile/$something_here/photos
	// but NOT /profile/photos neither /profile/$somethinng_here
	// and /profile/$something_here/photos/ (if PathCorrection is false).
	mux.HandleFunc("/profile/:name/photos", profilePhotosHandler)

	// wildcard, matches everything else after /uploads or /uploads/,
	// the param value of the "file" is all the path segments but without the first slash.
	mux.HandleFunc("/uploads/*file", listUploadsHandler)

	// named parameter in the same prefix as our previous registered wildcard
	// followed by static part (yes, this is also possible here!),
	// this has a priority over the /uploads/*file,
	// so if only /uploads/$something_here without other path segments
	// then it will fire the below handler:
	mux.HandleFunc("/uploads/:uploader", func(w http.ResponseWriter, r *http.Request) {
		uploader := muxie.GetParam(w, "uploader")
		fmt.Fprintf(w, "Hello Uploader: '%s'", uploader)
	})

	// static part followed by another wildcard in the same prefix as our previous registered wildcard,
	// and... yes, it is possible when you use muxie!
	mux.HandleFunc("/uploads/info/*file", func(w http.ResponseWriter, r *http.Request) {
		file := muxie.GetParam(w, "file")
		fmt.Fprintf(w, "File info of: '%s'", file)
	})

	// static part in the same path prefix as our previous registered wildcard and named parameter
	// (yes, this ia also possible here!).
	// This has priority over everyhing else after /uploads and /uploads/ (if PathCorrection is true).
	mux.HandleFunc("/uploads/totalsize", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Uploads total size is 4048")
	})

	// At the last 4 routes you see that this customized trie-based mux
	// has noumerous features and it is fast in the same time, if not the fastest!
	// You also learnt that you can use the "closest wildcard resolution" (/path/*myparam)
	// to do actions like custom 404 pages if nothing else found,
	// you can use it as root wildcard as well (/*myparam).
	// Navigate to the next example to learn how you can add your own 404 not found handler.

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf8")
	fmt.Fprintf(w, "This is the <strong>%s</strong>", "index page")
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	name := muxie.GetParam(w, "name")
	fmt.Fprintf(w, "Profile of: '%s'", name)
}

func profilePhotosHandler(w http.ResponseWriter, r *http.Request) {
	name := muxie.GetParam(w, "name")
	fmt.Fprintf(w, "Photos of: '%s'", name)
}

func listUploadsHandler(w http.ResponseWriter, r *http.Request) {
	file := muxie.GetParam(w, "file")
	fmt.Fprintf(w, "Showing file: '%s'", file)
}
