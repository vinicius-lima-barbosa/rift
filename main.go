package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Exercise: Create a HTTP inspector. `GET /inspect` should return:
// {
//   "method": "GET",
//   "path": "/inspect",
//   "protocol": "HTTP/1.1",
//   "host": "localhost:8080",
//   "userAgent": "curl/..."
// }
// And add a `POST /echo` to the curl responde a "hello rift".

type InspectResponse struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	Host      string `json:"host"`
	UserAgent string `json:"user_agent"`
}

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := InspectResponse{
		Method:    r.Method,
		Path:      r.URL.Path,
		Protocol:  r.Proto,
		Host:      r.Host,
		UserAgent: r.UserAgent(),
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func ecoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if _, err := io.Copy(w, r.Body); err != nil {
		http.Error(w, "failed to copy request body", http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /inspect", inspectHandler)
	mux.HandleFunc("POST /echo", ecoHandler)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
