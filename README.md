<h1 align="center">Muxie</h1>

<div align="center">
  :steam_locomotive::train::train::train::train::train:
</div>
<div align="center">
  <strong>Fast trie implementation designed from scratch specifically for HTTP</strong>
</div>
<div align="center">
  A <code>small and light</code> router for creating sturdy backend <a href="https://golang.org" alt="Go Programming Language">Go</a> applications
</div>

<br />

<div align="center">
  <!-- Release -->
  <a href="https://github.com/kataras/muxie/releases">
    <img src="https://img.shields.io/badge/release%20-v1.0-0077b3.svg?style=flat-squaree"
      alt="Release/stability" />
  </a>
  <!-- Build Status -->
  <a href="https://travis-ci.org/kataras/muxie">
    <img src="https://img.shields.io/travis/kataras/muxie/master.svg?style=flat-square"
      alt="Build Status" />
  </a>
  <!-- Report Card -->
  <a href="https://goreportcard.com/report/github.com/kataras/muxie">
    <img src="https://img.shields.io/badge/report%20card-a%2B-ff3333.svg?style=flat-square"
      alt="Report Card" />
  </a>
  <!-- Examples -->
  <a href="https://github.com/kataras/muxie/tree/master/_examples">
    <img src="https://img.shields.io/badge/learn%20by-examples-yellow.svg?style=flat-square"
      alt="Example" />
  </a>
  <!-- Built for Iris -->
  <a href="https://github.com/kataras/iris">
    <img src="https://img.shields.io/badge/built%20for-iris-0077b3.svg?style=flat-square"
      alt="Built for Iris" />
  </a>
</div>

<div align="center">
  <sub>The little router that could. Built with ❤︎ by
  <a href="https://twitter.com/MakisMaropoulos">Gerasimos Maropoulos</a>
</div>

<!-- [![Benchmark chart between muxie, httprouter, gin, gorilla mux, echo, vestigo and chi](_benchmarks/chart-17-oct-2018.png)](_benchmarks)

_Last updated on October 17, 2018._ Click [here](_benchmarks/README.md) to read more details. -->

## Features

- __trie based:__ [performance](_benchmarks/README.md) and useness are first class citizens, Muxie is based on the prefix tree data structure, designed from scratch and built for HTTP, and it is among the fastest outhere, if not the fastest one
- __grouping:__ group common routes based on their path prefixes
- __no external dependencies:__ weighing `30kb`, Muxie is a tiny little library without external dependencies
- __closest wildcard resolution and prefix-based custom 404:__ wildcards, named parameters and static paths can all live and play together nice and fast in the same path prefix or suffix(!)
- __small api:__ with only 3 main methods for HTTP there's not much to learn
- __compatibility:__ built to be 100% compatible with the `net/http` standard package

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl/)

```sh
$ go get -u github.com/kataras/muxie
```

## Example

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/rs/cors"

    "github.com/kataras/muxie"
)

func main() {
    mux := muxie.NewMux()
    mux.PathCorrection = true

    // _examples/6_middleware
    mux.Use(cors.Default().Handler)

    mux.HandleFunc("/", indexHandler)
    // Root wildcards, can be used for site-level custom not founds(404).
    mux.HandleFunc("/*path", notFoundHandler)

    // Grouping.
    profile := mux.Of("/profile")
    profile.HandleFunc("/:name", profileHandler)
    profile.HandleFunc("/:name/photos", profilePhotosHandler)
    // Wildcards can be used for prefix-level custom not found handler as well,
    // order does not matter.
    profile.HandleFunc("/*path", profileNotFoundHandler)

    // Dynamic paths with named parameters and wildcards or all together!
    mux.HandleFunc("/uploads/*file", listUploadsHandler)

    mux.HandleFunc("/uploads/:uploader", func(w http.ResponseWriter, r *http.Request) {
        uploader := muxie.GetParam(w, "uploader")
        fmt.Fprintf(w, "Hello Uploader: '%s'", uploader)
    })

    mux.HandleFunc("/uploads/info/*file", func(w http.ResponseWriter, r *http.Request) {
        file := muxie.GetParam(w, "file")
        fmt.Fprintf(w, "File info of: '%s'", file)
    })

    mux.HandleFunc("/uploads/totalsize", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, "Uploads total size is 4048")
    })

    fmt.Println("Server started at :8080")
    http.ListenAndServe(":8080", mux)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
    requestPath := muxie.GetParam(w, "path")
    // or r.URL.Path, we are in the root so it doesn't really matter.

    fmt.Fprintf(w, "Global Site Page of: '%s' not found", requestPath)
}

func profileNotFoundHandler(w http.ResponseWriter, r *http.Request) {
    requestSubPath := muxie.GetParam(w, "path")
    // requestSubPath = everyhing else after "http://localhost:8080/profile/..." 
    // but not /profile/:name or /profile/:name/photos because those will
    // be handled by the above route handlers we registered previously.

    fmt.Fprintf(w, "Profile Page of: '%s' not found", requestSubPath)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html;charset=utf8")
    fmt.Fprintf(w, "This is the <strong>%s</strong>", "index page")
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    name := muxie.GetParam(w, "name")
    fmt.Fprintf(w, "Profile of: '%s'", name)
}

func profilePhotosHandler(w http.ResponseWriter, r *http.Request) {
    name := muxie.GetParam(w, "name")
    fmt.Fprintf(w, "Photos of: '%s'", name)
}

func listUploadsHandler(w http.ResponseWriter, r *http.Request) {
    file := muxie.GetParam(w, "file")
    fmt.Fprintf(w, "Showing file: '%s'", file)
}

```
Want to see more examples and documentation? Check out the [examples](_examples).

## Philosophy

I believe that providing the right tools for the right job represents my best self
and I really enjoy writing small libraries and even frameworks that can be used and learnt by thousands like me.
I do it for the past two and a half years and I couldn't be more happy and proud for myself.

[Iris](https://github.com/kataras/iris) is a web backend framework for Go that is well-known in the Go community,
some of you hated it due to a "battle" between "competitors" followed by a single article written almost three years ago but the majority of you really love it so much that you recommend it to your co-workers, use it inside your companies, startups or your client's projects or even write your postgraduate dissertation based on your own experience with Iris. Both categories of fans
gave me enough reasons to continue and overcome myself day by day.

It was about the first days of September(2018) that I decided to start working on the next Iris release(version 11) and all job interviews postponed indefinitely.
If you have ever seen or hear about Iris, you already know that Iris is one of the fastest and easy-to-use frameworks, this is why it became so famous in so little time after all. 

A lot improvements were pushed over a month+ working full-time on Iris.
I have never seen a router or a framework supports so many patterns as the current Iris' internal router that is exposed by a beautiful API. However, I couldn't release it for the public yet, I felt that something was missing, I believed that I could do its router smarter and even faster(!) and that ate my guts. And then...in early October, after a lot of testing(and drinking) I found the missing part, it was that the routes' parameterized paths, wildcards and statics all-together for the same path prefix cannot play as fast as possible and good as they should, also I realised that the internal router's code was not the best ever written (it was written to be extremely fast and I didn't care about readability so much back then, when I finally made it to work faster than the rest I forgot to "prettify" it due to my excitement!)

Initially the `trie.go` and `node.go` were written for the Iris web framework's version 11 as you can understand by now, I believe that programming should be fun and not stressful, especially for new Gophers. So here we are, introducing a new autonomous Go-based mux(router) that it is light, fast and easy to use for all Gophers, not just for Iris users/developers.

The `kataras/muxie` repository contains the full source code of my trie implementation and the HTTP component(`muxie.NewMux()`) which is fully compatible with the `net/http` package. Users of this package are not limited on HTTP, they can use it to store and search simple key-value data into their programs (`muxie.NewTrie()`).


- The trie implementation is easy to read, and if it is not for you please send me a message to explain to you personally
- The API is simple, just three main methods and the two of them are the well-known `Handle` and `HandleFunc`, identically to the std package's `net/http#ServeMux`
- Implements a way to store parameters without touching the `*http.Request` and change the standard handler definition by introducing a new type such as a Context or slow the whole HTTP serve process because of it, look the `GetParam` function and the internal `paramsWriter` structure that it is created and used inside the `Mux#ServeHTTP`
- Besides the HTTP main functionality that this package offers, users should be able to use it for other things as well, the API is exposed as much as you walk through to
- Supports named parameters and wildcards of course
- Supports static path segments(parts, nodes) and named parameters and wildcards for the same path prefix without losing a bit of performance, unlike others that by-design they can't even do it

For the hesitants: There is a [public repository](https://github.com/kataras/trie-examples-to-remember-again) (previously private) that you can follow the whole process of coding and designing until the final result of `kataras/muxie`'s.

And... never forget to put some fun in your life ❤︎

Yours,<br />
Gerasimos Maropoulos ([@MakisMaropoulos](https://twitter.com/MakisMaropoulos))

## License
[MIT](https://tldrlegal.com/license/mit-license)