package muxie

import (
	"net/http"
	"strings"
	"sync"
)

// Mux is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// nodes and calls the handler for the pattern that
// most closely matches the URL.
//
// Patterns name fixed, rooted paths and dynamic like /profile/:name
// or /profile/:name/friends or even /files/*file when ":name" and "*file"
// are the named parameters and wildcard parameters respectfully.
//
// Note that since a pattern ending in a slash names a rooted subtree,
// the pattern "/*myparam" matches all paths not matched by other registered
// patterns, but not the URL with Path == "/", for that you would need the pattern "/".
//
// See `NewMux`.
type Mux struct {
	// PathCorrection removes leading slashes from the request path.
	// Defaults to false, however is highly recommended to turn it on.
	PathCorrection bool
	// PathCorrectionNoRedirect if `PathCorrection` is set to true,
	// it will execute the handlers chain without redirection.
	// Defaults to false.
	PathCorrectionNoRedirect bool
	Routes                   *Trie

	paramsPool *sync.Pool

	// per mux
	root            string
	requestHandlers []RequestHandler
	beginHandlers   []Wrapper
}

// NewMux returns a new HTTP multiplexer which uses a fast, if not the fastest
// implementation of the trie data structure that is designed especially for path segments.
func NewMux() *Mux {
	return &Mux{
		Routes: NewTrie(),
		paramsPool: &sync.Pool{
			New: func() interface{} {
				return &paramsWriter{}
			},
		},
		root: "",
	}
}

// AddRequestHandler adds a full `RequestHandler` which is responsible
// to check if a handler should be executed via its `Matcher`,
// if the handler is executed
// then the router stops searching for this Mux' routes,
// RequestHandelrs have priority over the routes and the middlewares.
//
// The "requestHandler"'s Handler can be any http.Handler
// and a new `muxie.NewMux` as well. The new Mux will
// be not linked to this Mux by-default, if you want to share
// middlewares then you have to use the `muxie.Pre` to declare
// the shared middlewares and register them via the `Mux#Use` function.
func (m *Mux) AddRequestHandler(requestHandler RequestHandler) {
	m.requestHandlers = append(m.requestHandlers, requestHandler)
}

// HandleRequest adds a matcher and a (conditional) handler to be executed when "matcher" passed.
// If the "matcher" passed then the "handler" will be executed
// and this Mux' routes will be ignored.
//
// Look the `Mux#AddRequestHandler` for further details.
func (m *Mux) HandleRequest(matcher Matcher, handler http.Handler) {
	m.AddRequestHandler(&simpleRequestHandler{
		Matcher: matcher,
		Handler: handler,
	})
}

// Use adds middleware that should be called before each mux route's main handler.
// Should be called before `Handle/HandleFunc`. Order matters.
//
// A Wrapper is just a type of `func(http.Handler) http.Handler`
// which is a common type definition for net/http middlewares.
//
// To add a middleware for a specific route and not in the whole mux
// use the `Handle/HandleFunc` with the package-level `muxie.Pre` function instead.
// Functionality of `Use` is pretty self-explained but new gophers should
// take a look of the examples for further details.
func (m *Mux) Use(middlewares ...Wrapper) {
	m.beginHandlers = append(m.beginHandlers, middlewares...)
}

type (
	// Wrapper is just a type of `func(http.Handler) http.Handler`
	// which is a common type definition for net/http middlewares.
	Wrapper func(http.Handler) http.Handler

	// Wrappers contains `Wrapper`s that can be registered and used by a "main route handler".
	// Look the `Pre` and `For/ForFunc` functions too.
	Wrappers []Wrapper
)

// For registers the wrappers for a specific handler and returns a handler
// that can be passed via the `Handle` function.
func (w Wrappers) For(main http.Handler) http.Handler {
	if len(w) > 0 {
		for i, lidx := 0, len(w)-1; i <= lidx; i++ {
			main = w[lidx-i](main)
		}
	}

	return main
}

// ForFunc registers the wrappers for a specific raw handler function
// and returns a handler that can be passed via the `Handle` function.
func (w Wrappers) ForFunc(mainFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return w.For(http.HandlerFunc(mainFunc))
}

// Pre starts a chain of handlers for wrapping a "main route handler"
// the registered "middleware" will run before the main handler(see `Wrappers#For/ForFunc`).
//
// Usage:
// mux := muxie.NewMux()
// myMiddlewares :=  muxie.Pre(myFirstMiddleware, mySecondMiddleware)
// mux.Handle("/", myMiddlewares.ForFunc(myMainRouteHandler))
func Pre(middleware ...Wrapper) Wrappers {
	return Wrappers(middleware)
}

// Handle registers a route handler for a path pattern.
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.Routes.Insert(m.root+pattern,
		WithHandler(
			Pre(m.beginHandlers...).For(handler)))
}

// HandleFunc registers a route handler function for a path pattern.
func (m *Mux) HandleFunc(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	m.Handle(pattern, http.HandlerFunc(handlerFunc))
}

// ServeHTTP exposes and serves the registered routes.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, h := range m.requestHandlers {
		if h.Match(r) {
			h.ServeHTTP(w, r)
			return
		}
	}

	path := r.URL.Path

	if m.PathCorrection {
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			// Remove trailing slash and client-permanent rule for redirection,
			// if confgiuration allows that and path has an extra slash.

			// update the new path and redirect.
			// use Trim to ensure there is no open redirect due to two leading slashes
			r.URL.Path = pathSep + strings.Trim(path, pathSep)
			if !m.PathCorrectionNoRedirect {
				url := r.URL.String()
				method := r.Method
				// Fixes https://github.com/kataras/iris/issues/921
				// This is caused for security reasons, imagine a payment shop,
				// you can't just permantly redirect a POST request, so just 307 (RFC 7231, 6.4.7).
				if method == http.MethodPost || method == http.MethodPut {
					http.Redirect(w, r, url, http.StatusTemporaryRedirect)
					return
				}

				http.Redirect(w, r, url, http.StatusMovedPermanently)
				return
			}
		}
	}

	// r.URL.Query() is slow and will allocate a lot, although
	// the first idea was to not introduce a new type to the end-developers
	// so they are using this library as the std one, but we will have to do it
	// for the params, we keep that rule so a new ResponseWriter, which is an interface,
	// and it will be compatible with net/http will be introduced to store the params at least,
	// we don't want to add a third parameter or a global state to this library.

	pw := m.paramsPool.Get().(*paramsWriter)
	pw.reset(w)
	n := m.Routes.Search(path, pw)
	if n != nil {
		n.Handler.ServeHTTP(pw, r)
	} else {
		http.NotFound(w, r)
		// or...
		// http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		// w.WriteHeader(http.StatusNotFound)
		// doesn't matter because the end-dev can customize the 404 with a root wildcard ("/*path")
		// which will be fired if no other requested path's closest wildcard is found.
	}

	m.paramsPool.Put(pw)
}

// SubMux is the child of a main Mux.
type SubMux interface {
	Of(prefix string) SubMux
	Unlink() SubMux
	Use(middlewares ...Wrapper)
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handlerFunc func(http.ResponseWriter, *http.Request))
	AbsPath() string
}

// Of returns a new Mux which its Handle and HandleFunc will register the path based on given "prefix", i.e:
// mux := NewMux()
// v1 := mux.Of("/v1")
// v1.HandleFunc("/users", myHandler)
// The above will register the "myHandler" to the "/v1/users" path pattern.
func (m *Mux) Of(prefix string) SubMux {
	if prefix == "" || prefix == pathSep {
		return m
	}

	if prefix == m.root {
		return m
	}

	// modify prefix if it's already there on the parent.
	if strings.HasPrefix(m.root, prefix) {
		prefix = prefix[0:strings.LastIndex(m.root, prefix)]
	}

	// remove last slash "/", if any.
	if lidx := len(prefix) - 1; prefix[lidx] == pathSepB {
		prefix = prefix[0:lidx]
	}

	// remove any duplication of slashes "/".
	prefix = pathSep + strings.Trim(m.root+prefix, pathSep)

	return &Mux{
		Routes: m.Routes,

		root:            prefix,
		requestHandlers: m.requestHandlers[0:],
		beginHandlers:   m.beginHandlers[0:],
	}
}

// AbsPath returns the absolute path of the router for this Mux group.
func (m *Mux) AbsPath() string {
	if m.root == "" {
		return "/"
	}
	return m.root
}

/* Notes:

Four options to solve optionally "inherition" of parent's middlewares but dismissed:

- I could add options for "inherition" of middlewares inside the `Mux#Use` itself.
  But this is a problem because the end-dev will have to use a specific muxie's constant even if he doesn't care about the other option.
- Create a new function like `UseOnly` or `UseExplicit`
  which will remove any previous middlewares and use only the new one.
  But this has a problem of making the `Use` func to act differently and debugging will be a bit difficult if big app if called after the `UseOnly`.
- Add a new func for creating new groups to remove any inherited middlewares from the parent.
  But with this, we will have two functions for the same thing and users may be confused about this API design.
- Put the options to the existing `Of` function, and make them optionally by functional design of options.
  But this will make things ugly and may confuse users as well, there is a better way.

Solution: just add a function like `Unlink`
to remove any inherited fields (now and future feature requests), so we don't have
breaking changes and etc. This `Unlink`, which will return the same SubMux, it can be used like `v1 := mux.Of(..).Unlink()`
*/

// Unlink will remove any inheritance fields from the parent mux (and its parent)
// that are inherited with the `Of` function.
// Returns the current SubMux. Usage:
//
// mux := NewMux()
// mux.Use(myLoggerMiddleware)
// v1 := mux.Of("/v1").Unlink() // v1 will no longer have the "myLoggerMiddleware" or any Matchers.
// v1.HandleFunc("/users", myHandler)
func (m *Mux) Unlink() SubMux {
	m.requestHandlers = m.requestHandlers[0:0]
	m.beginHandlers = m.beginHandlers[0:0]

	return m
}
