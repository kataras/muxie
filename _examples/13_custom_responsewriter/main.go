package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.Use(RequestTime)
	mux.HandleFunc("/profile/:name", profileHandler)
	fmt.Println(`Server started at http://localhost:8080
Open your browser or any other HTTP Client and navigate to:
http://localhost:8080/profile/yourname`)

	http.ListenAndServe(":8080", mux)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	name := muxie.GetParam(w, "name")
	fmt.Fprintf(w, "Hello, %s!", name)
}

type responseWriterWithTimer struct {
	// muxie.ParamStore
	// http.ResponseWriter
	// OR
	*muxie.Writer
	// OR/and implement the ParamStore interface by your own if you want
	// to customize the way the parameters are stored and retrieved.
	isHeaderWritten bool
	start           time.Time
}

// RequestTime is a middleware which modifies the response writer to use the `responseWriterWithTimer`.
// Look at: https://github.com/kataras/muxie/issues/10 too.
func RequestTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.(muxie.ParamStore), w OR cast it directly to the *muxie.Writer:
		next.ServeHTTP(&responseWriterWithTimer{w.(*muxie.Writer), false, time.Now()}, r)
	})
}

func (w *responseWriterWithTimer) WriteHeader(statusCode int) {
	elapsed := time.Since(w.start)
	w.Header().Set("X-Response-Time", strconv.FormatInt(elapsed.Nanoseconds(), 10))

	w.ResponseWriter.WriteHeader(statusCode)
	w.isHeaderWritten = true
}

func (w *responseWriterWithTimer) Write(b []byte) (int, error) {
	if !w.isHeaderWritten {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(b)
}
