package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/users", listUsersWithoutMuxieMethods)

	mux.Handle("/user/:id", muxie.Methods().
		HandleFunc(http.MethodGet, getUser).
		HandleFunc(http.MethodPost, saveUser).
		HandleFunc(http.MethodDelete, deleteUser))
	/*
		muxie.Methods().
			HandleFunc("POST, PUT", func(w http.ResponseWriter, r *http.Request) {[...]}

		^ can accept many methods for the same handler
		^ methods should be separated by comma, comma following by a space or just space

		Equivalent to:
		mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request){
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				w.Header().Set("Allow", "POST, PUT")
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			[...]
		})
	*/

	log.Println("Server started at http://localhost:8080\nGET: http://localhost:8080/users\nGET, POST, DELETE: http://localhost:8080/user/:id")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// The `muxie.Methods()` is just a helper for this common matching.
//
// However, you definitely own your route handlers,
// therefore you can easly make these checks manually
// by matching the `r.Method`.
func listUsersWithoutMuxieMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, "GET: List all users\n")
}

func getUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "GET: User details by user ID: %s\n", muxie.GetParam(w, "id"))
}

func saveUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "POST: save user with ID: %s\n", muxie.GetParam(w, "id"))
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "DELETE: remove user with ID: %s\n", muxie.GetParam(w, "id"))
}
