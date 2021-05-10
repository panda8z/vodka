package main

import (
	"log"
	"net/http"

	"github.com/panda8z/vodka/vodka"
)

func main() {
	log.Fatal(http.ListenAndServe(":8989", &vodka.Engin{}))
}
