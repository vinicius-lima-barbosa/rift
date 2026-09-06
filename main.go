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

	// The Copy concept comes from Streaming.
	// If a 5GB Request is sent, the proxy doesn't gonna read everything and send to memory.
	// Instead, he'll work over parts of the data withou maintaning the entire content.

	// Caso uma requisição de 5GB seja enviada, o proxy não vai ler tudo e enviar para a memória.
	// Ele vai trabalhar sobre partes dos dados sem precisar manter o conteúdo inteiro.

	// To Test:
	// curl -i \
	// 	-X POST \
	// 	http://localhost:8080/echo \
	// 	-d 'hello rift'
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
