package main

import (
	"context"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		addr := ":" + cfg.Metrics.Port
		slog.Info("metrics server started", "addr", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			slog.Error("metrics server failed", "error", err)
		}
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting scraper",
		"store", cfg.Store.Title,
		"categories", len(lenta.KnownCategories),
	)

	start := time.Now()
	totalProducts := 0

	for result := range scraper.ScrapeAll(ctx, lenta.KnownCategories) {
		if result.Err != nil {
			slog.Error("category failed",
				"category", result.Category.Name,
				"error", result.Err,
			)
			continue
		}

		csvFile, err := export.SaveCSV(result.Products, result.Category.Slug)
		if err != nil {
			slog.Error("CSV export failed", "category", result.Category.Name, "error", err)
		} else {
			slog.Info("CSV saved", "file", csvFile, "products", len(result.Products))
		}

		totalProducts += len(result.Products)
	}

	slog.Info("scraping finished",
		"total_products", totalProducts,
		"duration", time.Since(start).Round(time.Millisecond),
	)
}
