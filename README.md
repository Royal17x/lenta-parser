# Lenta Parser

Production-grade price monitoring parser for [lenta.com](https://lenta.com) .

## What it does

Collects product names, prices and page URLs from lenta.com by category, for a specific store (pickup point). Data is extracted from Angular SSR Transfer State embedded in HTML - no fragile CSS selectors, works regardless of markup changes.

## Stack

**Backend:** Go, Prometheus, `log/slog`, cleanenv  
**Frontend:** React 18, TypeScript, Tailwind CSS  
**Infra:** Docker Compose, Prometheus, Grafana, GitHub Actions

## Key technical decisions

- **HTML SSR parsing instead of XHR API** — lenta.com embeds product data in Angular Transfer State inside the HTML response. This is more stable than scraping DOM or intercepting private API endpoints.
- **Proxy rotation** — each request goes through a random proxy from the pool to avoid rate limiting.
- **Exponential backoff with jitter** — prevents thundering herd on retries.
- **Prometheus metrics** — `products_scraped_total`, `scrape_errors_total`, `request_duration_seconds`, `active_requests`.
- **SSE for real-time progress** — frontend receives live updates during parsing via Server-Sent Events.

## Metrics

| Metric | Type | Labels |
|---|---|---|
| `lenta_products_scraped_total` | Counter | `category` |
| `lenta_scrape_errors_total` | Counter | `category`, `reason` |
| `lenta_http_request_duration_seconds` | Histogram | `category` |
| `lenta_active_requests` | Gauge | — |

## Quick start

```bash
# Clone
git clone https://github.com/Royal17x/lenta-parser
cd lenta-parser

# Run backend
cd backend
go run ./cmd/parser

# Or with Docker Compose (coming in Sprint 6)
docker compose up
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `STORE_ID` | `3149` | Lenta store ID (ТК124, Мозаика) |
| `STORE_ALIAS` | `0124` | Store alias |
| `STORE_CITY` | `moscow` | City slug |
| `HTTP_TIMEOUT` | `30s` | Request timeout |
| `HTTP_RETRY_COUNT` | `3` | Max retry attempts |
| `HTTP_PROXIES` | `""` | Comma-separated proxy list |
| `SERVER_PORT` | `8080` | HTTP server port |
| `METRICS_PORT` | `9090` | Prometheus metrics port |

## Project structure
```
lenta-parser/
├── backend/
│   ├── cmd/parser/        # entrypoint
│   └── internal/
│       ├── config/        # env config
│       ├── client/        # HTTP client, proxy, retry
│       ├── lenta/         # parser, models, categories, cookies
│       ├── metrics/       # Prometheus metrics
│       ├── export/        # CSV/JSON export
│       └── server/        # HTTP server, SSE
├── frontend/              # React dashboard (coming soon)
└── infra/                 # Docker, Prometheus, Grafana (coming soon)
```