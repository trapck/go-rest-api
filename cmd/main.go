package main

import (
	"log"
	"net/http"

	"github.com/trapck/go-rest-api/server"
)

func main() {
	handler := http.HandlerFunc(server.BlogServer)
	if err := http.ListenAndServe(":3000", handler); err != nil {
		log.Fatalf("could not listen on port 3000 %v", err)
	}
}
