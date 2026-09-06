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

	// Execute and see the output:
	// curl -v \
	// 	-H 'X-Rift-Test: hello' \
	// 	http://localhost:8080/debug

	mux.HandleFunc("GET /debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Method: %s\n", r.Method)
		fmt.Fprintf(w, "Path: %s\n", r.URL.Path)
		fmt.Fprintf(w, "Host: %s\n", r.Host)
		fmt.Fprintf(w, "Protocol: %s\n", r.Proto)
		fmt.Fprintf(w, "Header: %s\n", r.Header.Get("X-Rift-Test"))
	})

	log.Println("Server listening on :8080") // Default PORT

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
