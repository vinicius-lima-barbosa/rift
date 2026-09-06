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
	return r.Method == rule.Method &&
		r.URL.Path == rule.Path
}

type RuleEngine struct {
	Rules []Rule
}

func (engine RuleEngine) Match(r *http.Request) (Rule, bool) {
	for _, rule := range engine.Rules {
		if rule.Match(r) {
			return rule, true
		}
	}

	return Rule{}, false
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

func latencyMiddleware(engine RuleEngine, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rule, matched := engine.Match(r)

		if !matched {
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

	engine := RuleEngine{
		Rules: []Rule{
			{
				Method: http.MethodPost,
				Path:   "/payments",
				Delay:  500 * time.Millisecond,
			},
			{
				Method: http.MethodGet,
				Path:   "/users",
				Delay:  200 * time.Millisecond,
			},
			{
				Method: http.MethodDelete,
				Path:   "/orders",
				Delay:  time.Second,
			},
			{
				Method: http.MethodPost,
				Path:   "/payments",
				Delay:  2 * time.Second,
			},
		},
	}

	handlerWithLatency := latencyMiddleware(engine, handler)
	handlerWithTiming := timingMiddleware(handlerWithLatency)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithTiming); err != nil {
		log.Fatal(err)
	}
}
