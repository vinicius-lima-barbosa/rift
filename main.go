package main

import (
	"fmt"
	"log"
	"net/http"
)

type Handler struct{}

func (h Handler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "%s %s", r.Method, r.URL.Path) // Client Response
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path) // Server
		next.ServeHTTP(w, r)
	})
}

func main() {
	handler := Handler{}

	handlerWithLogging := loggingMiddleware(handler)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithLogging); err != nil {
		log.Fatal(err)
	}
}
