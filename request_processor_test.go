package muxie

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type person struct {
	XMLName     xml.Name `json:"-" xml:"person"`
	Name        string   `json:"name" xml:"name,attr"`
	Age         int      `json:"age" xml:"age,attr"`
	Description string   `json:"description" xml:"description"`
}

func testProcessor(t *testing.T, p Processor, cType, tmplStrValue string) {
	testValue := person{Name: "kataras", Age: 25, Description: "software engineer"}
	testValueStr := fmt.Sprintf(tmplStrValue, testValue.Name, testValue.Age, testValue.Description)

	mux := NewMux()
	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		var v person
		if err := Bind(r, p, &v); err != nil {
			t.Fatal(err)
		}

		if expected, got := v.Name, testValue.Name; expected != got {
			t.Fatalf("expected name to be: '%s' but got: '%s'", expected, got)
		}
		if expected, got := v.Age, testValue.Age; expected != got {
			t.Fatalf("expected age to be: '%d' but got: '%d'", expected, got)
		}

		if expected, got := v.Description, testValue.Description; expected != got {
			t.Fatalf("expected description to be: '%s' but got: '%s'", expected, got)
		}
	})
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		if err := Dispatch(w, p, testValue); err != nil {
			t.Fatal(err)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	expectWithBody(t, http.MethodGet, srv.URL+"/read", testValueStr,
		http.Header{"Content-Type": []string{cType}}).statusCode(http.StatusOK)

	expect(t, http.MethodGet, srv.URL+"/write").statusCode(http.StatusOK).
		headerEq("Content-Type", withCharset(cType)).
		bodyEq(testValueStr)
}

func TestJSON(t *testing.T) {
	testProcessor(t, JSON, "application/json", `{"name":"%s","age":%d,"description":"%s"}`)
}

func TestXML(t *testing.T) {
	testProcessor(t, XML, "text/xml", `<person name="%s" age="%d"><description>%s</description></person>`)
}
