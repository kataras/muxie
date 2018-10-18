package muxie

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetParam(t *testing.T) {
	mux := NewMux()

	mux.HandleFunc("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		name := GetParam(w, "name")
		fmt.Fprintf(w, "Hello %s", name)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	expect(t, http.MethodGet, srv.URL+"/hello/kataras").bodyEq("Hello kataras")
}
