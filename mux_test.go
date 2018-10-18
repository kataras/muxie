package muxie

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func expect(t *testing.T, method, url string) *testie {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return &testie{t: t, res: res}
}

type testie struct {
	t   *testing.T
	res *http.Response
}

func (te *testie) statusCode(expected int) *testie {
	if got := te.res.StatusCode; expected != got {
		te.t.Fatalf("%s: expected status code: %d but got %d", te.res.Request.URL, expected, got)
	}

	return te
}

func (te *testie) bodyEq(expected string) *testie {
	b, err := ioutil.ReadAll(te.res.Body)
	te.res.Body.Close()
	if err != nil {
		te.t.Fatal(err)
	}

	if got := string(b); expected != got {
		te.t.Fatalf("%s: expected to receive '%s' but got '%s'", te.res.Request.URL, expected, got)
	}

	return te
}

func (te *testie) headerEq(key, expected string) *testie {
	if got := te.res.Header.Get(key); expected != got {
		te.t.Fatalf("%s: expected header value of %s to be: '%s' but got '%s'", te.res.Request.URL, key, expected, got)
	}

	return te
}

func TestMuxPathCorrection(t *testing.T) {
	mux := NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/hello/here", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %s", r.URL.Query().Get("name"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	expect(t, http.MethodGet, srv.URL+"/hello//here/?name=kataras").bodyEq("Hello kataras")
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

	expect(t, http.MethodGet, srv.URL).bodyEq("Handler of /")
	expect(t, http.MethodGet, srv.URL+"/v1").bodyEq("Handler of /v1")
	expect(t, http.MethodGet, srv.URL+"/v1/hello").bodyEq("Handler of /v1/hello")
}
