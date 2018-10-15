package muxie

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxPathCorrection(t *testing.T) {
	mux := NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/hello/here", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %s", r.URL.Query().Get("name"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/hello//here/?name=kataras")
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

func TestMuxOf(t *testing.T) {
	printPathHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Handler of %s", r.URL.Path)
	}

	mux := NewMux()
	mux.HandleFunc("/", printPathHandler)

	v1 := mux.Of("/v1") // or "/v1/" or even "v1"
	v1.HandleFunc("/", printPathHandler)
	v1.HandleFunc("/hello", printPathHandler)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	res, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if expected, got := "Handler of /", string(greeting); expected != got {
		t.Fatalf("expected to receive '%s' but got '%s'", expected, got)
	}

	res, err = http.Get(srv.URL + "/v1/hello")
	if err != nil {
		t.Fatal(err)
	}

	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if expected, got := "Handler of /v1/hello", string(greeting); expected != got {
		t.Fatalf("expected to receive '%s' but got '%s'", expected, got)
	}

	res, err = http.Get(srv.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}

	greeting, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if expected, got := "Handler of /v1", string(greeting); expected != got {
		t.Fatalf("expected to receive '%s' but got '%s'", expected, got)
	}
}
