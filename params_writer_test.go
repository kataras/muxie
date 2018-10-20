package muxie

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGetParam(t *testing.T) {
	mux := NewMux()

	mux.HandleFunc("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		name := GetParam(w, "name")
		fmt.Fprintf(w, "Hello %s", name)
	})

	testHandler(t, mux, http.MethodGet, "/hello/kataras").bodyEq("Hello kataras")
}
