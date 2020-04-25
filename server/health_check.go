package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
)

func (s *Server) healthCheck(ctx context.Context) http.Handler {
	h := &healthChecker{s: s}

	h.runHealthCheck()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 30):
			}
			h.runHealthCheck()
		}
	}()
	return h
}

// healthChecker periodically performs health checks on server dependencies.
// Currently, it only checks that the storage layer is available.
type healthChecker struct {
	s *Server

	// Result of the last health check: any error and the amount of time it took
	// to query the storage.
	mu sync.RWMutex
	// Guarded by the mutex
	err    error
	passed time.Duration
}

// runHealthCheck performs a single health check and makes the result available
// for any clients performing and HTTP request against the healthChecker.
func (h *healthChecker) runHealthCheck() {
	t := time.Now()
	err := checkStorageHealth(h.s.store)
	passed := time.Since(t)
	if err != nil {
		log.Errorf("server: storage health check failed: %s", err)
	}

	// Make sure to only hold the mutex to access the fields, and not while
	// we're querying the storage object.
	h.mu.Lock()
	defer h.mu.Unlock()
	h.err = err
	h.passed = passed
}

func checkStorageHealth(s sessions.Store) error {
	req, _ := http.NewRequest("GET", "http://www.example.com", nil)
	session, _ := s.New(req, "health-check")
	rsp := httptest.NewRecorder()
	err := session.Save(req, rsp)
	if err != nil {
		return fmt.Errorf("save session error: %v", err)
	}

	session.Options.MaxAge = -1
	err = session.Save(req, rsp)
	if err != nil {
		return fmt.Errorf("remove session error: %v", err)
	}

	return nil
}

func (h *healthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	err := h.err
	t := h.passed
	h.mu.RUnlock()

	if err != nil {
		http.Error(w, "health check failed", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(w, "Health check passed in %s", t)
}
