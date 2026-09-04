package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// In this simple logic we have: TCP Connection -> HTTP Server -> Parse HTTP -> HTTP Request -> Handler -> Response Writer -> HTTP Response
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	log.Println("Server listening on :8080") // Default PORT

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
