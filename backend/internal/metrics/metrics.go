package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProductsScraped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lenta_products_scraped_total",
		Help: "Total number of products successfully scraped",
	}, []string{"category"})

	ScrapeErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lenta_scrape_errors_total",
		Help: "Total number of scrape errors",
	}, []string{"category", "reason"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lenta_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"category"})

	ActiveRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lenta_active_requests",
		Help: "Number of currently active HTTP requests",
	})
)
