package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte("Welcome!\n"))
	})

	r.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte(mux.Vars(r)["id"]))
	})

	fmt.Println("Server started at localhost:3000")
	http.ListenAndServe(":3000", r)
}
