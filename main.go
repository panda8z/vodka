package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

type Engin struct{}

func (e *Engin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", html.EscapeString(req.URL.Path))
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func main() {
	log.Fatal(http.ListenAndServe(":8989", &Engin{}))
}
