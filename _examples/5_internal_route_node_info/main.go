package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/index", indexHandler)
	mux.HandleFunc("/about", aboutHandler)

	v1 := mux.Of("/v1")
	v1.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "List of all users...")
	})

	v1.HandleFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id := getParamUint64(w, "id", 0)
		if id == 0 {
			http.Error(w, "invalid user id", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Details of user with ID: %d", id)
	})

	// So far all good, nothing new shown above,
	// let's see how we can get the registered endpoints based on a prefix or a root node,
	// for more see the `muxie.Mux#Routes` godocs.
	//
	//
	// request: http://localhost:8080/nodes/v1
	// response:
	// /v1/users
	// /v1/users/:id
	nodesRouter := mux.Of("/nodes")
	nodesRouter.HandleFunc("/v1", func(w http.ResponseWriter, r *http.Request) {
		v1Node := mux.Routes.SearchPrefix("/v1")

		if v1Node == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		v1ChildNodeKeys := v1Node.Keys(nil)
		for _, key := range v1ChildNodeKeys {
			fmt.Fprintln(w, key)
		}
	})

	// http://localhost:8080
	// http://localhost:8080/index
	// http://localhost:8080/about
	// http://localhost:8080/v1/users
	// http://localhost:8080/v1/users/42
	// http://localhost:8080/nodes/v1
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf8")
	fmt.Fprintf(w, "This is the <strong>%s</strong>", "index page")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("About Page\n"))
}

// getParamUint64 returns the param's value as uint64.
// If not found returns "def".
func getParamUint64(w http.ResponseWriter, key string, def uint64) uint64 {
	v := muxie.GetParam(w, key)
	if v == "" {
		return def
	}

	val, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return def
	}
	if val > math.MaxUint64 {
		return def
	}

	return val
}
