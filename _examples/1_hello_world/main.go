package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()

	// if it is true the /about/ will be permantly redirected to /about and served from the aboutHandler.
	// mux.PathCorrection = true

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/index", indexHandler)
	mux.HandleFunc("/about", aboutHandler)

	fmt.Println(`Server started at :8080
Open your browser or any other HTTP Client and navigate to:
http://localhost:8080
http://localhost:8080/index and
http://localhost:8080/about`)

	http.ListenAndServe(":8080", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf8")
	fmt.Fprintf(w, "This is the <strong>%s</strong>", "index page")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Simple example to show how easy is to add routes with static paths.\nVisit the 'parameterized' example folder for more...\n"))
}
