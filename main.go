package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Rule types

type RequestMatch struct {
	Method string
	Path   string
}

type LatencyFault struct {
	Delay time.Duration
}

type AbortFault struct {
	StatusCode int
	Message    string
}

func (request RequestMatch) Match(r *http.Request) bool {
	return r.Method == request.Method &&
		r.URL.Path == request.Path
}

// Rule

type Rule struct {
	RequestMatch RequestMatch
	Latency      *LatencyFault
	Abort        *AbortFault
}

type RuleEngine struct {
	Rules []Rule
}

func (engine RuleEngine) Match(r *http.Request) (Rule, bool) {
	for _, rule := range engine.Rules {
		if rule.RequestMatch.Match(r) {
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

func faultMiddleware(engine RuleEngine, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rule, matched := engine.Match(r)

		if !matched {
			next.ServeHTTP(w, r)
			return
		}

		if rule.Latency != nil {
			if rule.Abort != nil {
				log.Printf("fault=abort method=%s path=%s\n", r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(rule.Abort.StatusCode)

				fmt.Fprintln(w, rule.Abort.Message)
				return
			}

			log.Printf("fault=latency method=%s path=%s delay=%s\n", r.Method, r.URL.Path, rule.Latency.Delay)

			timer := time.NewTimer(rule.Latency.Delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				log.Println("server responds")
				next.ServeHTTP(w, r)
			case <-ctx.Done():
				log.Printf("request canceled method=%s path=%s error=%v\n", r.Method, r.URL.Path, ctx.Err())
				return
			}
		} else {
			if rule.Abort != nil {
				log.Printf("fault=abort method=%s path=%s\n", r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(rule.Abort.StatusCode)

				fmt.Fprintln(w, rule.Abort.Message)
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}

func main() {
	handler := Handler{}

	engine := RuleEngine{
		Rules: []Rule{
			{
				RequestMatch: RequestMatch{
					Method: http.MethodPost,
					Path:   "/payments",
				},
				Latency: &LatencyFault{
					Delay: 500 * time.Millisecond,
				},
			},
			{
				RequestMatch: RequestMatch{
					Method: http.MethodGet,
					Path:   "/users",
				},
				Abort: &AbortFault{
					StatusCode: http.StatusServiceUnavailable,
					Message:    "service unavailable",
				},
			},
		},
	}

	handlerWithLatency := faultMiddleware(engine, handler)
	handlerWithTiming := timingMiddleware(handlerWithLatency)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithTiming); err != nil {
		log.Fatal(err)
	}
}
