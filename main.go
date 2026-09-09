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

func (match RequestMatch) Match(r *http.Request) bool {
	return r.Method == match.Method &&
		r.URL.Path == match.Path
}

func (match RequestMatch) Validate() error {
	if match.Method == "" {
		return fmt.Errorf("method is required")
	}

	if match.Path == "" {
		return fmt.Errorf("path is required")
	}

	if match.Path[0] != '/' {
		return fmt.Errorf("path must start with '/'")
	}

	return nil
}

// Fault

type Fault interface {
	Apply(
		w http.ResponseWriter,
		r *http.Request,
		next http.Handler,
	)
	Validate() error
}

func (fault LatencyFault) Apply(
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
) {
	log.Printf(
		"fault=latency method=%s path=%s delay=%s",
		r.Method,
		r.URL.Path,
		fault.Delay,
	)

	timer := time.NewTimer(fault.Delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		next.ServeHTTP(w, r)

	case <-r.Context().Done():
		log.Printf(
			"request canceled method=%s path=%s error=%v",
			r.Method,
			r.URL.Path,
			r.Context().Err(),
		)
	}
}

func (fault AbortFault) Apply(
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
) {
	log.Printf(
		"fault=abort method=%s path=%s status=%d",
		r.Method,
		r.URL.Path,
		fault.StatusCode,
	)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(fault.StatusCode)

	fmt.Fprintln(w, fault.Message)
}

func (fault LatencyFault) Validate() error {
	if !(fault.Delay > 0) {
		return fmt.Errorf("latency delay must be greater than zero")
	}

	return nil
}

func (fault AbortFault) Validate() error {
	if fault.StatusCode < 400 || fault.StatusCode > 599 {
		return fmt.Errorf("status should be higher than 400 or lower than 599")
	}

	if fault.Message == "" {
		return fmt.Errorf("message should not be empty")
	}

	return nil
}

// Rule

type Rule struct {
	RequestMatch RequestMatch
	Fault        Fault
}

type RuleEngine struct {
	Rules []Rule
}

func (rule Rule) Validate() error {
	if err := rule.RequestMatch.Validate(); err != nil {
		return err
	}

	if rule.Fault == nil {
		return fmt.Errorf("fault is required")
	}

	if err := rule.Fault.Validate(); err != nil {
		return err
	}

	return nil
}

func (engine RuleEngine) Match(r *http.Request) (Rule, bool) {
	for _, rule := range engine.Rules {
		if rule.RequestMatch.Match(r) {
			return rule, true
		}
	}

	return Rule{}, false
}

func (engine RuleEngine) Validate() error {
	for i, rule := range engine.Rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return nil
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

func faultMiddleware(
	engine RuleEngine,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		rule, matched := engine.Match(r)

		if !matched {
			next.ServeHTTP(w, r)
			return
		}

		rule.Fault.Apply(w, r, next)
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
			},
			{
				RequestMatch: RequestMatch{
					Method: http.MethodGet,
					Path:   "/users",
				},
				Fault: AbortFault{
					StatusCode: http.StatusServiceUnavailable,
					Message:    "service unavailable",
				},
			},
		},
	}

	if err := engine.Validate(); err != nil {
		log.Fatal(err)
	}

	handlerWithFaults := faultMiddleware(engine, handler)
	handlerWithTiming := timingMiddleware(handlerWithFaults)

	log.Println("server listening on 8080")

	if err := http.ListenAndServe(":8080", handlerWithTiming); err != nil {
		log.Fatal(err)
	}
}
