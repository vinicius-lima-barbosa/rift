package main

import (
	"encoding/json"
	"fmt"
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

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/inspect", inspectHandler)
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello rift")
	})

	log.Println("Server listening on 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
