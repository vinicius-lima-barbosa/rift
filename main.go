package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Handler struct{}

func (h Handler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "%s %s\n", r.Method, r.URL.Path) // Client Response
}

func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s duration=%s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func latencyMiddleware(delay time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		timer := time.NewTimer(delay)
		defer timer.Stop()

		log.Println("lantecy middleware started")

		select {
		case <-timer.C:
			log.Println("server responds")
			next.ServeHTTP(w, r)
		case <-ctx.Done():
			log.Printf("request canceled: %v\n", ctx.Err())
			return
		}
	})
}

func main() {
	handler := Handler{}

	handlerWithLatency := latencyMiddleware(500*time.Millisecond, handler)
	handlerWithTiming := timingMiddleware(handlerWithLatency)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithTiming); err != nil {
		log.Fatal(err)
	}
}
