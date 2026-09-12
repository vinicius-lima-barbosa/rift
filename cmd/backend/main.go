package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	upstream := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler o corpo", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		fmt.Fprintf(
			w,
			"method: %s\npath: %s\nquery: %s\ncontent-type: %s\nx-rift-test: %s\nbody: %s\n",
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			r.Header.Get("Content-Type"),
			r.Header.Get("X-Rift-Test"), // Captura o header customizado
			string(bodyBytes),           // Converte os bytes do JSON para string
		)
	})

	log.Println("server listening on 9000")

	if err := http.ListenAndServe(":9000", upstream); err != nil {
		log.Fatal(err)
	}
}
