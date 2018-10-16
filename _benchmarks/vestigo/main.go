package main

import (
	"fmt"
	"net/http"

	"github.com/husobee/vestigo"
)

func main() {
	r := vestigo.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome!\n"))
	})

	r.Get("/user/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(vestigo.Param(r, "id")))
	})

	fmt.Println("Server started at localhost:3000")
	http.ListenAndServe(":3000", r)
}
