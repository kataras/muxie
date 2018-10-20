package muxie

import (
	"bytes"
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

	return testReq(t, req)
}

func expectWithBody(t *testing.T, method, url string, body string, headers http.Header) *testie {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}

	if len(headers) > 0 {
		req.Header = http.Header{}
		for k, v := range headers {
			req.Header[k] = v
		}
	}

	return testReq(t, req)
}

func testReq(t *testing.T, req *http.Request) *testie {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	resp.Request = req
	return &testie{t: t, resp: resp}
}

func testHandler(t *testing.T, handler http.Handler, method, url string) *testie {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, nil)
	handler.ServeHTTP(w, req)
	resp := w.Result()
	resp.Request = req
	return &testie{t: t, resp: resp}
}

type testie struct {
	t    *testing.T
	resp *http.Response
}

func (te *testie) statusCode(expected int) *testie {
	if got := te.resp.StatusCode; expected != got {
		te.t.Fatalf("%s: expected status code: %d but got %d", te.resp.Request.URL, expected, got)
	}

	return te
}

func (te *testie) bodyEq(expected string) *testie {
	b, err := ioutil.ReadAll(te.resp.Body)
	te.resp.Body.Close()
	if err != nil {
		te.t.Fatal(err)
	}

	if got := string(b); expected != got {
		te.t.Fatalf("%s: expected to receive '%s' but got '%s'", te.resp.Request.URL, expected, got)
	}

	return te
}

func (te *testie) headerEq(key, expected string) *testie {
	if got := te.resp.Header.Get(key); expected != got {
		te.t.Fatalf("%s: expected header value of %s to be: '%s' but got '%s'", te.resp.Request.URL, key, expected, got)
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
