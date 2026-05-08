package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Royal17x/lenta-parser/internal/export"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Royal17x/lenta-parser/internal/lenta"
)

type Server struct {
	scraper  Scraper
	maxPages int

	mu      sync.RWMutex
	results []lenta.ScrapeResult
	running bool
	events  []sseEvent
}

type sseEvent struct {
	Type     string `json:"type"`
	Category string `json:"category,omitempty"`
	Message  string `json:"message,omitempty"`
	Count    int    `json:"count,omitempty"`
}

type Scraper interface {
	ScrapeAll(ctx context.Context, categories []lenta.Category, maxPages int) <-chan lenta.ScrapeResult
}

func New(scraper Scraper, maxPages int) *Server {
	return &Server{scraper: scraper, maxPages: maxPages}
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
	s.events = nil
	s.mu.Unlock()

	go s.runScraping()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) runScraping() {
	ctx := context.Background()
	slog.Info("scraping started via API")

	for result := range s.scraper.ScrapeAll(ctx, lenta.KnownCategories, s.maxPages) {
		s.mu.Lock()
		s.results = append(s.results, result)
		if result.Err != nil {
			s.events = append(s.events, sseEvent{
				Type: "error", Category: result.Category.Name,
				Message: result.Err.Error(),
			})
		} else {
			s.events = append(s.events, sseEvent{
				Type: "progress", Category: result.Category.Name,
				Count: len(result.Products),
			})
		}
		s.mu.Unlock()
		if result.Err == nil {
			if file, err := export.SaveCSV(result.Products, result.Category.Slug); err != nil {
				slog.Error("CSV export failed", "category", result.Category.Name, "error", err)
			} else {
				slog.Info("CSV saved", "file", file)
			}
		}
	}

	s.mu.Lock()
	total := 0
	for _, r := range s.results {
		total += len(r.Products)
	}
	s.events = append(s.events, sseEvent{
		Type: "done", Count: total,
		Message: fmt.Sprintf("Собрано %d товаров", total),
	})
	s.running = false
	s.mu.Unlock()

	slog.Info("scraping finished via API", "total", total)
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

	send := func(ev sseEvent) {
		data, _ := json.Marshal(ev)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	sent := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			all := s.events
			running := s.running
			s.mu.RUnlock()

			for sent < len(all) {
				send(all[sent])
				sent++
				if all[sent-1].Type == "done" {
					return
				}
			}

			if !running && sent == len(all) && sent > 0 && all[sent-1].Type == "done" {
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
