package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/panda8z/vodka/vodka"
)

func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", html.EscapeString(req.URL.Path))
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

func main() {
	engine := vodka.New()
	engine.GET("/", indexHandler)
	engine.GET("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8989", engine))
}
