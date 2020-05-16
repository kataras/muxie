package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	mux.Use(corsMiddleware) // <--- IMPORTANT: register the cors middleware.

	mux.Handle("/", muxie.Methods().
		NoContent(http.MethodOptions). // <--- IMPORTANT: cors preflight.
		HandleFunc(http.MethodPost, postHandler))

	fmt.Println("Server started at http://localhost:80")
	http.ListenAndServe(":80", mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			h.Set("Access-Control-Methods", "POST, PUT, PATCH, DELETE")
			h.Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,Content-Type")
			h.Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var request map[string]interface{}
	muxie.JSON.Bind(r, &request)
	muxie.JSON.Dispatch(w, map[string]string{"message": "ok"})
}
