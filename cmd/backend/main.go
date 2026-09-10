package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	upstream := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		fmt.Fprintf(
			w,
			"UPSTREAM: %s %s\n",
			r.Method,
			r.URL.Path,
		)
	})

	log.Println("server listening on 9000")

	if err := http.ListenAndServe(":9000", upstream); err != nil {
		log.Fatal(err)
	}
}
