package main

import (
	"fmt"
	"net/http"

	"github.com/kataras/muxie"
)

func main() {
	mux := muxie.NewMux()
	mux.HandleFunc("/", handleRootDomainIndex)

	// mysubdomain
	mySubdomain := muxie.NewMux()
	mySubdomain.HandleFunc("/", handleMySubdomainIndex)
	mySubdomain.HandleFunc("/about", aboutHandler)

	mux.Match(muxie.Host("mysubdomain.localhost:8080"), mySubdomain)
	//

	// mysubsubdomain.mysubdomain
	mySubSubdomain := muxie.NewMux()
	mySubSubdomain.HandleFunc("/", handleMySubSubdomainIndex)
	mySubSubdomain.HandleFunc("/about", aboutHandler)

	mux.Match(muxie.Host("mysubsubdomain.mysubdomain.localhost:8080"), mySubSubdomain)
	//

	// any other subdomain
	myWildcardSubdomain := muxie.NewMux()
	myWildcardSubdomain.HandleFunc("/", handleMyWildcardSubdomainIndex)
	// Catch any other host that ends with .localhost:8080.
	mux.Match(muxie.Host(".localhost:8080"), myWildcardSubdomain)
	/*
		Or add a custom match func that validates if the router
		should proceed with this subdomain handler:
		This one is extremely useful for apps that may need dynamic subdomains based on a database,
		usernames for example.
		mux.MatchFunc(func(r *http.Request) bool{
			return userRepo.Exists(...use of http.Request)
		}, myWildcardSubdomain)
		Or
		mux.AddMatcher(_value_of_a_struct_which_completes_the_muxie.MatcherHandler)
	*/

	// Chrome-based browsers will automatically work but to test with
	// firefox or a custom http client or POSTMAN you may want to edit your hosts,
	// i.e on windows is going like this:
	// 127.0.0.1 mysubdomain.localhost
	// 127.0.0.1 mysubsubdomain.mysubdomain.localhost
	//
	// You may run your own virtual domain if you change the listening addr ":8080"
	// to something like "mydomain.com:80".
	//
	// Read more at godocs of `Mux#AddMatcher`.
	fmt.Println(`Server started at http://localhost:8080
Open your browser and navigate through:
http://mysubdomain.localhost:8080
http://mysubdomain.localhost:8080/about
http://mysubsubdomain.mysubdomain.localhost:8080
http://mysubsubdomain.mysubdomain.localhost:8080/about
http://any.subdomain.can.be.handled.by.asterix.localhost:8080`)
	http.ListenAndServe(":8080", mux)
}

func handleRootDomainIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[0] Hello from the root domain: %s\n", r.Host)
}

func handleMySubdomainIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[1] Hello from mysubdomain.localhost:8080\n")
}

func handleMySubSubdomainIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[2] Hello from mysubsubdomain.mysubdomain.localhost:8080\n")
}

func handleMyWildcardSubdomainIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[3] I can handle any subdomain's index page / if non of the statics found, so hello from host: %s\n", r.Host)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "About of: %s\n", r.Host)
}
