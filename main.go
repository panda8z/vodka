package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func fooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}
func main() {
	http.HandleFunc("/bar", fooHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
