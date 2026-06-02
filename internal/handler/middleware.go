// Module: github.com/yourname/code-snippet-explainer
// RateLimiter is an HTTP middleware that limits requests per IP address.
//
// NewRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler
//   - limit:  max requests allowed per IP within the window (use 10)
//   - window: rolling time window (use 1 * time.Minute)
//
// Implementation:
//   - Use a sync.Mutex-protected map[string]*ipState to track per-IP counts.
//   - ipState holds: Count int, WindowStart time.Time
//   - On each request, get the client IP from r.RemoteAddr (strip port).
//   - If the window has expired, reset Count and WindowStart.
//   - If Count >= limit, return 429 JSON: {"error":"rate limit exceeded, try again in N seconds"}
//     where N is seconds until the window resets.
//   - Otherwise increment Count and call next.ServeHTTP.
//   - Run a cleanup goroutine every 5 minutes to delete stale IP entries.

package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ipState struct {
	Count       int
	WindowStart time.Time
}

func NewRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	ipMap := make(map[string]*ipState)
	var mu sync.Mutex
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, state := range ipMap {
				if time.Since(state.WindowStart) > window {
					delete(ipMap, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if colon := len(ip) - 1 - len(":65535"); colon > 0 && ip[colon] == ':' {
				ip = ip[:colon]
			}
			mu.Lock()
			state, exists := ipMap[ip]
			now := time.Now()
			if !exists || now.Sub(state.WindowStart) > window {
				ipMap[ip] = &ipState{Count: 1, WindowStart: now}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}
			if state.Count >= limit {
				resetIn := window - now.Sub(state.WindowStart)
				mu.Unlock()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, "<p class=\"error\">Rate limit exceeded. Try again in %d seconds.</p>", int(resetIn.Seconds())+1)
				return
			}
			state.Count++
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
