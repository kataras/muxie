package muxie

import (
	"net/http"
	"strings"
)

type (
	// RequestHandler is the matcher and handler link interface.
	// It is used inside the `Mux` to handle requests based on end-developer's custom logic.
	// If a "Matcher" passed then the "Handler" is executing and the rest of the Mux' routes will be ignored.
	RequestHandler interface {
		http.Handler
		Matcher
	}

	simpleRequestHandler struct {
		http.Handler
		Matcher
	}

	// Matcher is the interface that all Matchers should be implemented
	// in order to be registered into the Mux via the `Mux#AddRequestHandler/Match/MatchFunc` functions.
	//
	// Look the `Mux#AddRequestHandler` for more.
	Matcher interface {
		Match(*http.Request) bool
	}

	// MatcherFunc is a shortcut of the Matcher, as a function.
	// See `Matcher`.
	MatcherFunc func(*http.Request) bool
)

// Match returns the result of the "fn" matcher.
// Implementing the `Matcher` interface.
func (fn MatcherFunc) Match(r *http.Request) bool {
	return fn(r)
}

// Host is a Matcher for hostlines.
// It can accept exact hosts line like "mysubdomain.localhost:8080"
// or a suffix, i.e ".localhost:8080" will work as a wildcard subdomain for our root domain.
// The domain and the port should match exactly the request's data.
type Host string

// Match validates the host, implementing the `Matcher` interface.
func (h Host) Match(r *http.Request) bool {
	s := string(h)
	return r.Host == s || (s[0] == '.' && strings.HasSuffix(r.Host, s)) || s == WildcardParamStart
}
