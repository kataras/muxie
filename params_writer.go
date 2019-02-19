package muxie

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// GetParam returns the path parameter value based on its key, i.e
// "/hello/:name", the parameter key is the "name".
// For example if a route with pattern of "/hello/:name" is inserted to the `Trie` or handlded by the `Mux`
// and the path "/hello/kataras" is requested through the `Mux#ServeHTTP -> Trie#Search`
// then the `GetParam("name")` will return the value of "kataras".
// If not associated value with that key is found then it will return an empty string.
//
// The function will do its job only if the given "w" http.ResponseWriter interface is an `ResponseWriter`.
func GetParam(w http.ResponseWriter, key string) string {
	if store, ok := w.(ResponseWriter); ok {
		return store.Get(key)
	}

	return ""
}

// GetParams returns all the available parameters based on the "w" http.ResponseWriter which should be a ResponseWriter.
//
// The function will do its job only if the given "w" http.ResponseWriter interface is an `ResponseWriter`.
func GetParams(w http.ResponseWriter) []ParamEntry {
	if store, ok := w.(ResponseWriter); ok {
		return store.GetAll()
	}

	return nil
}

// SetParam sets manually a parameter to the "w" http.ResponseWriter which should be a ResponseWriter.
// This is not commonly used by the end-developers,
// unless sharing values(string messages only) between handlers is absolutely necessary.
func SetParam(w http.ResponseWriter, key, value string) bool {
	if store, ok := w.(ResponseWriter); ok {
		store.Set(key, value)
		return true
	}

	return false
}

// ParamEntry holds the Key and the Value of a named path parameter.
type ParamEntry struct {
	Key   string
	Value string
}

// ResponseWriter is the muxie's specific ResponseWriter to hold the path parameters.
// Usage: use this to cast a handler's `http.ResponseWriter` and pass it as an embedded parameter to custom response writer
// that will be passed to the next handler in the chain.
type ResponseWriter interface {
	http.ResponseWriter
	ParamsSetter
	Get(string) string
	GetAll() []ParamEntry
	// Status returns the status code of the response or 0 if the response has
	// not been written
	Status() int
	// Written returns whether or not the ResponseWriter has been written.
	Written() bool
	// Size returns the size of the response body.
	Size() int
	// Before allows for a function to be called before the ResponseWriter has been written to. This is
	// useful for setting headers or any other operations that must happen before a response has been written.
	Before(func(ResponseWriter))
}

type beforeFunc func(ResponseWriter)

type paramsWriter struct {
	http.ResponseWriter
	params      []ParamEntry
	status      int
	size        int
	beforeFuncs []beforeFunc
}

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(pw http.ResponseWriter) ResponseWriter {
	npw := &paramsWriter{
		ResponseWriter: pw,
	}

	if _, ok := pw.(http.CloseNotifier); ok {
		return &responseWriterCloseNotifer{npw}
	}

	return npw
}

var _ ResponseWriter = (*paramsWriter)(nil)

// Set implements the `ParamsSetter` which `Trie#Search` needs to store the parameters, if any.
// These are decoupled because end-developers may want to use the trie to design a new Mux of their own
// or to store different kind of data inside it.
func (pw *paramsWriter) Set(key, value string) {
	if ln := len(pw.params); cap(pw.params) > ln {
		pw.params = pw.params[:ln+1]
		p := &pw.params[ln]
		p.Key = key
		p.Value = value
		return
	}

	pw.params = append(pw.params, ParamEntry{
		Key:   key,
		Value: value,
	})
}

func (pw *paramsWriter) WriteHeader(s int) {
	pw.status = s
	pw.ResponseWriter.WriteHeader(s)
}

func (pw *paramsWriter) Write(b []byte) (int, error) {
	if !pw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		pw.WriteHeader(http.StatusOK)
	}
	size, err := pw.ResponseWriter.Write(b)
	pw.size += size
	return size, err
}

func (pw *paramsWriter) Flush() {
	flusher, ok := pw.ResponseWriter.(http.Flusher)
	if ok {
		if !pw.Written() {
			// The status will be StatusOK if WriteHeader has not been called yet
			pw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

func (pw *paramsWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := pw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (pw *paramsWriter) Status() int {
	return pw.status
}

func (pw *paramsWriter) Size() int {
	return pw.size
}

func (pw *paramsWriter) Written() bool {
	return pw.status != 0
}

// Get returns the value of the associated parameter based on its key/name.
func (pw *paramsWriter) Get(key string) string {
	n := len(pw.params)
	for i := 0; i < n; i++ {
		if kv := pw.params[i]; kv.Key == key {
			return kv.Value
		}
	}

	return ""
}

func (pw *paramsWriter) Before(before func(ResponseWriter)) {
	pw.beforeFuncs = append(pw.beforeFuncs, before)
}

func (pw *paramsWriter) callBefore() {
	for i := len(pw.beforeFuncs) - 1; i >= 0; i-- {
		pw.beforeFuncs[i](pw)
	}
}

// GetAll returns all the path parameters keys-values.
func (pw *paramsWriter) GetAll() []ParamEntry {
	return pw.params
}

func (pw *paramsWriter) reset(w http.ResponseWriter) {
	pw.ResponseWriter = w
	pw.params = pw.params[0:0]
}

type responseWriterCloseNotifer struct {
	*paramsWriter
}

func (pw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return pw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
