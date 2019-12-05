package muxie

import (
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
}

type paramsWriter struct {
	http.ResponseWriter
	params []ParamEntry
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

// GetAll returns all the path parameters keys-values.
func (pw *paramsWriter) GetAll() []ParamEntry {
	return pw.params
}

func (pw *paramsWriter) reset(w http.ResponseWriter) {
	pw.ResponseWriter = w
	pw.params = pw.params[0:0]
}

// Flusher indicates if `Flush` is supported by the client.
//
// The default HTTP/1.x and HTTP/2 ResponseWriter implementations
// support Flusher, but ResponseWriter wrappers may not. Handlers
// should always test for this ability at runtime.
//
// Note that even for ResponseWriters that support Flush,
// if the client is connected through an HTTP proxy,
// the buffered data may not reach the client until the response
// completes.
func (pw *paramsWriter) Flusher() (http.Flusher, bool) {
	flusher, canFlush := pw.ResponseWriter.(http.Flusher)
	return flusher, canFlush
}

// Flush sends any buffered data to the client.
func (pw *paramsWriter) Flush() {
	if flusher, ok := pw.Flusher(); ok {
		flusher.Flush()
	}
}
