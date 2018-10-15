package muxie

import (
	"fmt"
	"io/ioutil"
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

	res, err := http.Get(srv.URL + "/hello/kataras")
	if err != nil {
		t.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if expected, got := "Hello kataras", string(greeting); expected != got {
		t.Fatalf("expected to receive '%s' but got '%s'", expected, got)
	}
}
