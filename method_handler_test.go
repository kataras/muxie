package muxie

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethodHandler(t *testing.T) {
	mux := NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprintf(w, "GET: List all users\n")
	})

	mux.Handle("/user/:id", Methods().
		HandleFunc(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "GET: User details by user ID: %s\n", GetParam(w, "id"))
		}).
		HandleFunc(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "POST: save user with ID: %s\n", GetParam(w, "id"))
		}).
		HandleFunc(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "DELETE: remove user with ID: %s\n", GetParam(w, "id"))
		}))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	expect(t, http.MethodGet, srv.URL+"/users").statusCode(http.StatusOK).
		bodyEq("GET: List all users\n")
	expect(t, http.MethodPost, srv.URL+"/users").statusCode(http.StatusMethodNotAllowed).
		bodyEq("Method Not Allowed\n").headerEq("Allow", "GET")

	expect(t, http.MethodGet, srv.URL+"/user/42").statusCode(http.StatusOK).
		bodyEq("GET: User details by user ID: 42\n")
	expect(t, http.MethodPost, srv.URL+"/user/42").statusCode(http.StatusOK).
		bodyEq("POST: save user with ID: 42\n")
	expect(t, http.MethodDelete, srv.URL+"/user/42").statusCode(http.StatusOK).
		bodyEq("DELETE: remove user with ID: 42\n")
	expect(t, http.MethodPut, srv.URL+"/user/42").statusCode(http.StatusMethodNotAllowed).
		bodyEq("Method Not Allowed\n").headerEq("Allow", "GET, POST, DELETE")
}
