package main

import (
	"fmt"
	"log"
	"net/http"
)

type RiftHandler struct{}

func (h *RiftHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Rift intercepted %s %s\n", r.Method, r.URL.Path)
}

func main() {
	handler := RiftHandler{}

	log.Println("Server listening on 8080")

	if err := http.ListenAndServe(":8080", &handler); err != nil {
		log.Fatal(err)
	}
}
