package muxie

import (
	"net/http"
	"testing"
)

func TestRequestHandler(t *testing.T) {
	const (
		domain                    = "localhost.suff"
		domainResponse            = "Hello from root domain"
		subdomain                 = "mysubdomain." + domain
		subdomainResponse         = "Hello from " + subdomain
		subdomainAboutResposne    = "About the " + subdomain
		wildcardSubdomain         = "." + domain
		wildcardSubdomainResponse = "Catch all subdomains"

		customMethod = "CUSTOM"
	)

	mux := NewMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(domainResponse))
	})

	subdomainHandler := NewMux() // can have its own request handlers as well.
	subdomainHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(subdomainResponse))
	})
	subdomainHandler.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(subdomainAboutResposne))
	})

	mux.HandleRequest(Host(subdomain), subdomainHandler)
	mux.HandleRequest(Host(wildcardSubdomain), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(wildcardSubdomainResponse))
	}))

	mux.HandleRequest(MatcherFunc(func(r *http.Request) bool {
		return r.Method == customMethod
	}), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(customMethod))
	}))

	testHandler(t, mux, http.MethodGet, "http://"+domain).
		statusCode(http.StatusOK).bodyEq(domainResponse)

	testHandler(t, mux, http.MethodGet, "http://"+subdomain).
		statusCode(http.StatusOK).bodyEq(subdomainResponse)

	testHandler(t, mux, http.MethodGet, "http://"+subdomain+"/about").
		statusCode(http.StatusOK).bodyEq(subdomainAboutResposne)

	testHandler(t, mux, http.MethodGet, "http://anysubdomain.here.for.test"+subdomain).
		statusCode(http.StatusOK).bodyEq(wildcardSubdomainResponse)

	testHandler(t, mux, customMethod, "http://"+domain).
		statusCode(http.StatusOK).bodyEq(customMethod)
}
