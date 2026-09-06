package main

import (
	"fmt"
	"log"
	"net/http"
)

type InspectorHandler struct{}

func (h InspectorHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "%s %s", r.Method, r.URL.Path)
}

func main() {
	handler := InspectorHandler{}

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
