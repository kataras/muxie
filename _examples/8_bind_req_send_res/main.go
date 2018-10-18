package main

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

// Read more at https://golang.org/pkg/encoding/xml
type person struct {
	XMLName     xml.Name `json:"-" xml:"person"`                // element name
	Name        string   `json:"name" xml:"name,attr"`          // ,attr for attribute.
	Age         int      `json:"age" xml:"age,attr"`            // ,attr attribute.
	Description string   `json:"description" xml:"description"` // inner element name, value is its body.
}

func main() {
	mux := muxie.NewMux()
	mux.PathCorrection = true

	// Read from incoming request.
	mux.Handle("/save", muxie.Methods().
		HandleFunc("POST, PUT", func(w http.ResponseWriter, r *http.Request) {
			var p person
			// muxie.Bind(r, muxie.JSON,...) for JSON.
			// You can implement your own Binders by implementing the muxie.Binder interface like the muxie.JSON/XML.
			err := muxie.Bind(r, muxie.XML, &p)
			if err != nil {
				http.Error(w, fmt.Sprintf("unable to read body: %v", err), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Go value of the request body:\n%#+v\n", p)
		}))

	// Send a response.
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		send := person{Name: "kataras", Age: 25, Description: "software engineer"}

		// muxie.Dispatch(w, muxie.JSON,...) for JSON.
		// You can implement your own Dispatchers by implementing the muxie.Dispatcher interface like the muxie.JSON/XML.
		err := muxie.Dispatch(w, muxie.XML, send)
		if err != nil {

			http.Error(w, fmt.Sprintf("unable to send the value of %#+v. Error: %v", send, err), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println(`Server started at http://localhost:8080
:: How to...
Read from incoming request

request:
    POST or PUT: http://localhost:8080/save
request body:
    <person name="kataras" age="25"><description>software engineer</description></person>
request header:
    "Content-Type": "text/xml"
response:
    Go value of the request body:
    main.person{XMLName:xml.Name{Space:"", Local:"person"}, Name:"kataras", Age:25, Description:"software engineer"}

Send a response

request:
    GET: http://localhost:8080/get
response header:
    "Content-Type": "text/xml; charset=utf-8" (can be modified by muxie.Charset variable)
response:
    <person name="kataras" age="25">
        <description>software engineer</description>
    </person>`)
	http.ListenAndServe(":8080", mux)
}
