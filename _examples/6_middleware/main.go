// Package main will explore the helpers for middleware(s) that Muxie has to offer,
// but they are totally optional, you can still use your favourite pattern
// to wrap route handlers.
//
// Example of usage of an external net/http common middleware:
//
// import "github.com/rs/cors"
//
// mux := muxie.New()
// mux.Use(cors.Default().Handler)
//
//
// To wrap a specific route or even if for some reason you want to wrap the entire router
// use the `Pre(middlewares...).For(mainHandler)` as :
//
// wrapped := muxie.Pre(cors.Default().Handler, ...).For(mux)
// http.ListenAndServe(..., wrapped)
package main

import (
	"log"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	// Globally, will be inherited by all sub muxes as well unless `Of(...).Unlink()` called.
	mux.Use(myGlobalMiddleware)

	// Per Route.
	mux.Handle("/", muxie.Pre(myFirstRouteMiddleware, mySecondRouteMiddleware).ForFunc(myMainRouteHandler))

	// Per Group.
	inheritor := mux.Of("/inheritor")
	inheritor.Use(myMiddlewareForSubmux)
	inheritor.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: my inheritor's main index route's handler")
		w.Write([]byte("Hello from /inheritor\n"))
	})

	// Per Group, without its parents' middlewares.
	// Unlink will clear all middlewares for this sub mux.
	orphan := mux.Of("/orphan").Unlink()
	orphan.Use(myMiddlewareForSubmux)
	orphan.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: orphan's main index route's handler")
		w.Write([]byte("Hello from /orphan\n"))
	})

	// Open your web browser or any other HTTP Client
	// and navigate through the below endpoinds, one by one,
	// and check the console output of your webserver.
	//
	// http://localhost:8080
	// http://localhost:8080/inheritor
	// http://localhost:8080/orphan
	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func myGlobalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: my global and first of all middleware for all following mux' routes and sub muxes")
		next.ServeHTTP(w, r)
	})
}

func myMiddlewareForSubmux(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: my submux' routes middleware")
		next.ServeHTTP(w, r)
	})
}

func myFirstRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: my first specific route's middleware")
		next.ServeHTTP(w, r)
	})
}

func mySecondRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("execute: my second specific route's middleware")
		next.ServeHTTP(w, r)
	})
}

func myMainRouteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("execute: my main route's handler")
	w.Write([]byte("Hello World!\n"))
}
