package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Rule

type Rule struct {
	Method string
	Path   string
	Delay  time.Duration
}

func (rule Rule) Match(r *http.Request) bool {
	return r.Method == rule.Method && r.URL.Path == rule.Path
}

// Handler

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

func latencyMiddleware(rule Rule, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if !rule.Match(r) {
			next.ServeHTTP(w, r)
			return
		}

		timer := time.NewTimer(rule.Delay)
		defer timer.Stop()

		log.Printf("fault=latency method=%s path=%s delay=%s\n", r.Method, r.URL.Path, rule.Delay)

		select {
		case <-timer.C:
			log.Println("server responds")
			next.ServeHTTP(w, r)
		case <-ctx.Done():
			log.Printf("request canceled method=%s path=%s error=%v\n", r.Method, r.URL.Path, ctx.Err())
			return
		}
	})
}

func main() {
	handler := Handler{}

	rule := Rule{
		Method: http.MethodPost,
		Path:   "/payments",
		Delay:  500 * time.Millisecond,
	}

	handlerWithLatency := latencyMiddleware(rule, handler)
	handlerWithTiming := timingMiddleware(handlerWithLatency)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithTiming); err != nil {
		log.Fatal(err)
	}
}
