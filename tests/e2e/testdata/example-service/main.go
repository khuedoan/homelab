package main

import (
	"fmt"
	"log"
	"net/http"
)

var version = "development"

func main() {
	http.HandleFunc("/info", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, version)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
