package muxie

import (
	"net/http"
	"strings"
)

// Methods returns a MethodHandler which caller can use
// to register handler for specific HTTP Methods inside the `Mux#Handle/HandleFunc`.
// Usage:
// mux := muxie.NewMux()
// mux.Handle("/user/:id", muxie.Methods().
//     Handle("GET", getUserHandler).
//     Handle("POST", saveUserHandler))
func Methods() *MethodHandler {
	//
	// Design notes, the latest one is selected:
	//
	// mux := muxie.NewMux()
	//
	// 1. mux.Handle("/user/:id", muxie.ByMethod("GET", getHandler).And/AndFunc("POST", postHandlerFunc))
	//
	// 2. mux.Handle("/user/:id", muxie.ByMethods{
	// 	  "GET": getHandler,
	// 	  "POST" http.HandlerFunc(postHandlerFunc),
	//   }) <- the only downside of this is that
	// we lose the "Allow" header, which is not so important but it is RCF so we have to follow it.
	//
	// 3. mux.Handle("/user/:id", muxie.Method("GET", getUserHandler).Method("POST", saveUserHandler))
	//
	// 4. mux.Handle("/user/:id", muxie.Methods().
	//      Handle("GET", getHandler).
	//      HandleFunc("POST", postHandler))
	//
	return &MethodHandler{handlers: make(map[string]http.Handler)}
}

// NoContentHandler defaults to a handler which just sends 204 status.
// See `MethodHandler.NoContent` method.
var NoContentHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
})

// MethodHandler implements the `http.Handler` which can be used on `Mux#Handle/HandleFunc`
// to declare handlers responsible for specific HTTP method(s).
//
// Look `Handle` and `HandleFunc`.
type MethodHandler struct {
	// origin *Mux

	handlers map[string]http.Handler // method:handler
	// Handle/HandleFunc can accept more than one methods per handler separated by comma or space,
	// however in order to not repeat ourselves for every handler:
	// extra methods will be not registered to a handler but they can register
	// the route so it can be reachable, it is binded to a handler which just sends status no content,
	// can be used for OPTIONS on cors.
	noContentMethods []string

	methodsAllowedStr string
}

// Handle adds a handler to be responsible for a specific HTTP Method.
// Returns this MethodHandler for further calls.
// Usage:
// Handle("GET", myGetHandler).HandleFunc("DELETE", func(w http.ResponseWriter, r *http.Request){[...]})
// Handle("POST, PUT", saveOrUpdateHandler)
//        ^ can accept many methods for the same handler
//        ^ methods should be separated by comma, comma following by a space or just space
func (m *MethodHandler) Handle(method string, handler http.Handler) *MethodHandler {
	multiMethods := strings.FieldsFunc(method, func(c rune) bool {
		return c == ',' || c == ' '
	})

	if len(multiMethods) > 1 {
		for _, method := range multiMethods {
			m.Handle(method, handler)
		}

		return m
	}

	method = normalizeMethod(method)

	if m.methodsAllowedStr == "" {
		m.methodsAllowedStr = method
	} else {
		m.methodsAllowedStr += ", " + method
	}

	m.handlers[method] = handler

	return m
}

// NoContent registers a handler to a method
// which sends 204 (no status content) to the client.
//
// Example: _examples/11_cors for more.
func (m *MethodHandler) NoContent(methods ...string) *MethodHandler {
	for _, method := range methods {
		m.handlers[normalizeMethod(method)] = NoContentHandler
	}

	return m
}

// HandleFunc adds a handler function to be responsible for a specific HTTP Method.
// Returns this MethodHandler for further calls.
func (m *MethodHandler) HandleFunc(method string, handlerFunc func(w http.ResponseWriter, r *http.Request)) *MethodHandler {
	m.Handle(method, http.HandlerFunc(handlerFunc))
	return m
}

func (m *MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := m.handlers[r.Method]; ok {
		handler.ServeHTTP(w, r)
		return
	}

	// RCF rfc2616 https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html
	// The response MUST include an Allow header containing a list of valid methods for the requested resource.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Allow#Examples
	w.Header().Set("Allow", m.methodsAllowedStr)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func normalizeMethod(method string) string {
	return strings.ToUpper(strings.TrimSpace(method))
}
