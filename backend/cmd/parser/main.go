package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Royal17x/lenta-parser/internal/client"
	"github.com/Royal17x/lenta-parser/internal/config"
	"github.com/Royal17x/lenta-parser/internal/lenta"
	"github.com/Royal17x/lenta-parser/internal/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		slog.Info("metrics server started", "addr", ":"+cfg.Metrics.Port)
		http.ListenAndServe(":"+cfg.Metrics.Port, mux)
	}()

	httpClient, err := client.New(&cfg.HTTP)
	if err != nil {
		slog.Error("failed to create HTTP client", "error", err)
		os.Exit(1)
	}

	storeCfg := lenta.StoreConfig{
		ID:    cfg.Store.ID,
		Alias: cfg.Store.Alias,
		Title: cfg.Store.Title,
		City:  cfg.Store.City,
	}
	scraper := lenta.NewScraper(httpClient, storeCfg, cfg.HTTP.QratorJSID)
	srv := server.New(scraper, cfg.Scraper.MaxPagesPerCategory)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: srv.Routes(),
	}

	go func() {
		slog.Info("HTTP server started", "addr", ":"+cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down")
	httpServer.Shutdown(context.Background())
}
