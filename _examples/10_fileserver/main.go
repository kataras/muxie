package main

import (
	"log"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.Handle("/static/*file", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Println("Server started at http://localhost:8080\nGET: http://localhost:8080/static/\nGET: http://localhost:8080/static/js/empty.js")
	http.ListenAndServe(":8080", mux)
}
