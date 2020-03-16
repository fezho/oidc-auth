package server

import (
	"context"
	"net/http"
	"sync"
	"time"
)

func (s *Server) healthCheck(ctx context.Context) http.Handler {
	return nil
}

// healthChecker periodically performs health checks on server dependencies.
// Currently, it only checks that the storage layer is available.
type healthChecker struct {
	mu sync.RWMutex

	err    error
	passed time.Duration
}

// TODO: dex health check
