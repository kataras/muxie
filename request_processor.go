package muxie

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

var (
	// Charset is the default content type charset for Request Processors .
	Charset = "utf-8"

	// JSON implements the full `Processor` interface.
	// It is responsible to dispatch JSON results to the client and to read JSON
	// data from the request body.
	//
	// Usage:
	// To read from a request:
	// muxie.Bind(r, muxie.JSON, &myStructValue)
	// To send a response:
	// muxie.Dispatch(w, muxie.JSON, mySendDataValue)
	JSON = &jsonProcessor{Prefix: nil, Indent: "", UnescapeHTML: false}

	// XML implements the full `Processor` interface.
	// It is responsible to dispatch XML results to the client and to read XML
	// data from the request body.
	//
	// Usage:
	// To read from a request:
	// muxie.Bind(r, muxie.XML, &myStructValue)
	// To send a response:
	// muxie.Dispatch(w, muxie.XML, mySendDataValue)
	XML = &xmlProcessor{Indent: ""}
)

func withCharset(cType string) string {
	return cType + "; charset=" + Charset
}

// Binder is the interface which `muxie.Bind` expects.
// It is used to bind a request to a go struct value (ptr).
type Binder interface {
	Bind(*http.Request, interface{}) error
}

// Bind accepts the current request and any `Binder` to bind
// the request data to the "ptrOut".
func Bind(r *http.Request, b Binder, ptrOut interface{}) error {
	return b.Bind(r, ptrOut)
}

// Dispatcher is the interface which `muxie.Dispatch` expects.
// It is used to send a response based on a go struct value.
type Dispatcher interface {
	// no io.Writer because we need to set the headers here,
	// Binder and Processor are only for HTTP.
	Dispatch(http.ResponseWriter, interface{}) error
}

// Dispatch accepts the current response writer and any `Dispatcher`
// to send the "v" to the client.
func Dispatch(w http.ResponseWriter, d Dispatcher, v interface{}) error {
	return d.Dispatch(w, v)
}

// Processor implements both `Binder` and `Dispatcher` interfaces.
// It is used for implementations that can `Bind` and `Dispatch`
// the same data form.
//
// Look `JSON` and `XML` for more.
type Processor interface {
	Binder
	Dispatcher
}

var (
	newLineB byte = '\n'
	// the html codes for unescaping
	ltHex = []byte("\\u003c")
	lt    = []byte("<")

	gtHex = []byte("\\u003e")
	gt    = []byte(">")

	andHex = []byte("\\u0026")
	and    = []byte("&")
)

type jsonProcessor struct {
	Prefix       []byte
	Indent       string
	UnescapeHTML bool
}

var _ Processor = (*jsonProcessor)(nil)

func (p *jsonProcessor) Bind(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (p *jsonProcessor) Dispatch(w http.ResponseWriter, v interface{}) error {
	var (
		result []byte
		err    error
	)

	if indent := p.Indent; indent != "" {
		marshalIndent := json.MarshalIndent

		result, err = marshalIndent(v, "", indent)
		result = append(result, newLineB)
	} else {
		marshal := json.Marshal
		result, err = marshal(v)
	}

	if err != nil {
		return err
	}

	if p.UnescapeHTML {
		result = bytes.Replace(result, ltHex, lt, -1)
		result = bytes.Replace(result, gtHex, gt, -1)
		result = bytes.Replace(result, andHex, and, -1)
	}

	if len(p.Prefix) > 0 {
		result = append([]byte(p.Prefix), result...)
	}

	w.Header().Set("Content-Type", withCharset("application/json"))
	_, err = w.Write(result)
	return err
}

type xmlProcessor struct {
	Indent string
}

var _ Processor = (*xmlProcessor)(nil)

func (p *xmlProcessor) Bind(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return xml.Unmarshal(b, v)
}

func (p *xmlProcessor) Dispatch(w http.ResponseWriter, v interface{}) error {
	var (
		result []byte
		err    error
	)

	if indent := p.Indent; indent != "" {
		marshalIndent := xml.MarshalIndent

		result, err = marshalIndent(v, "", indent)
		result = append(result, newLineB)
	} else {
		marshal := xml.Marshal
		result, err = marshal(v)
	}

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", withCharset("text/xml"))
	_, err = w.Write(result)
	return err
}
