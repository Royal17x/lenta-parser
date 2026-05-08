package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Royal17x/lenta-parser/internal/lenta"
)

type Server struct {
	scraper Scraper
	mu      sync.RWMutex
	results []lenta.ScrapeResult
	running bool
}

type Scraper interface {
	ScrapeAll(ctx context.Context, categories []lenta.Category, maxPages int) <-chan lenta.ScrapeResult
}

func New(scraper Scraper) *Server {
	return &Server{scraper: scraper}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/start", s.handleStart)
	mux.HandleFunc("GET /api/stream", s.handleStream)
	mux.HandleFunc("GET /api/results", s.handleResults)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	return withCORS(mux)
}

type SSEEvent struct {
	Type     string `json:"type"`
	Category string `json:"category,omitempty"`
	Message  string `json:"message,omitempty"`
	Count    int    `json:"count,omitempty"`
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		http.Error(w, "already running", http.StatusConflict)
		return
	}
	s.running = true
	s.results = nil
	s.mu.Unlock()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	sendEvent := func(event SSEEvent) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	for {
		s.mu.RLock()
		running := s.running
		s.mu.RUnlock()
		if running {
			break
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(500 * time.Millisecond):
		}
	}

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	lastSent := 0
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			current := s.results
			running := s.running
			s.mu.RUnlock()

			for i := lastSent; i < len(current); i++ {
				res := current[i]
				if res.Err != nil {
					sendEvent(SSEEvent{
						Type:     "error",
						Category: res.Category.Name,
						Message:  res.Err.Error(),
					})
				} else {
					sendEvent(SSEEvent{
						Type:     "progress",
						Category: res.Category.Name,
						Count:    len(res.Products),
					})
				}
				lastSent++
			}

			if !running && lastSent >= len(current) {
				total := 0
				for _, r := range current {
					total += len(r.Products)
				}
				sendEvent(SSEEvent{
					Type:    "done",
					Count:   total,
					Message: fmt.Sprintf("Scraped %d products", total),
				})
				return
			}
		}
	}
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []lenta.Product
	for _, res := range s.results {
		all = append(all, res.Products...)
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(all)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, res := range s.results {
		total += len(res.Products)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"running":    s.running,
		"categories": len(s.results),
		"products":   total,
	})
}

func (s *Server) AppendResult(res lenta.ScrapeResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = append(s.results, res)
}

func (s *Server) SetRunning(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = v
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *Server) TotalProducts() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, r := range s.results {
		total += len(r.Products)
	}
	return total
}
