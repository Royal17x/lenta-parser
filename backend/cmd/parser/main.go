package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Royal17x/lenta-parser/internal/client"
	"github.com/Royal17x/lenta-parser/internal/config"
	"github.com/Royal17x/lenta-parser/internal/export"
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
	srv := server.New(scraper)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: srv.Routes(),
	}
	go func() {
		slog.Info("HTTP server started", "addr", ":"+cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting scraper",
		"store", cfg.Store.Title,
		"categories", len(lenta.KnownCategories),
		"max_pages", cfg.Scraper.MaxPagesPerCategory,
	)

	start := time.Now()
	srv.SetRunning(true)

	for result := range scraper.ScrapeAll(ctx, lenta.KnownCategories, cfg.Scraper.MaxPagesPerCategory) {
		srv.AppendResult(result)

		if result.Err != nil {
			slog.Error("category failed", "category", result.Category.Name, "error", result.Err)
			continue
		}

		csvFile, err := export.SaveCSV(result.Products, result.Category.Slug)
		if err != nil {
			slog.Error("CSV export failed", "error", err)
		} else {
			slog.Info("CSV saved", "file", csvFile, "products", len(result.Products))
		}
	}

	srv.SetRunning(false)
	slog.Info("scraping finished",
		"total_products", srv.TotalProducts(),
		"duration", time.Since(start).Round(time.Millisecond),
	)

	<-ctx.Done()
	httpServer.Shutdown(context.Background())
}
