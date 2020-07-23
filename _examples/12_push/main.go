// Server push lets the server preemptively "push" website assets
// to the client without the user having explicitly asked for them.
// When used with care, we can send what we know the user is going
// to need for the page they're requesting.
package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.HandleFunc("/", pushHandler)
	mux.HandleFunc("/main.js", simpleAssetHandler)

	http.ListenAndServeTLS(":443", "mycert.crt", "mykey.key", mux)
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	// The target must either be an absolute path (like "/path") or an absolute
	// URL that contains a valid host and the same scheme as the parent request.
	// If the target is a path, it will inherit the scheme and host of the
	// parent request.
	target := "/main.js"

	if pusher, ok := w.(*muxie.Writer).ResponseWriter.(http.Pusher); ok {
		err := pusher.Push(target, nil)
		if err != nil {
			if err == http.ErrNotSupported {
				http.Error(w, "HTTP/2 push not supported", http.StatusHTTPVersionNotSupported)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<html><body><script src="%s"></script></body></html>`, target)
}

func simpleAssetHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/main.js")
}
