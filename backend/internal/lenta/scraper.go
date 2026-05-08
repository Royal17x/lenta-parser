package lenta

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Royal17x/lenta-parser/internal/metrics"
)

const baseURL = "https://lenta.com"

type Scraper struct {
	client  HTTPClient
	cookies []*http.Cookie
}

type HTTPClient interface {
	Get(ctx context.Context, url string, category string, cookies []*http.Cookie) ([]byte, error)
}

func NewScraper(client HTTPClient, storeCfg StoreConfig, qratorJSID string) *Scraper {
	return &Scraper{
		client:  client,
		cookies: StoreCookiesWithQrator(storeCfg, qratorJSID),
	}
}

func (s *Scraper) ScrapeCategory(ctx context.Context, category Category, maxPages int) ([]Product, error) {
	url := fmt.Sprintf("%s/catalog/%s/", baseURL, category.Slug)

	slog.Info("scraping category", "name", category.Name, "page", 1)

	body, err := s.client.Get(ctx, url, category.Slug, s.cookies)
	if err != nil {
		metrics.ScrapeErrors.WithLabelValues(category.Slug, "http").Inc()
		return nil, fmt.Errorf("fetch page 1: %w", err)
	}

	products, err := ParseProductsFromHTML(body)
	if err != nil {
		metrics.ScrapeErrors.WithLabelValues(category.Slug, "parse").Inc()
		return nil, fmt.Errorf("parse page 1: %w", err)
	}

	total, _ := ParseTotalFromHTML(body)
	totalPages := (total + pageSize - 1) / pageSize
	if maxPages > 0 && totalPages > maxPages {
		totalPages = maxPages
	}

	slog.Info("category info", "name", category.Name, "total", total, "pages", totalPages)

	seen := make(map[int]struct{}, len(products))
	result := make([]Product, 0, totalPages*pageSize)
	for _, p := range products {
		seen[p.ID] = struct{}{}
		result = append(result, p)
	}

	for page := 2; page <= totalPages; page++ {
		select {
		case <-ctx.Done():
			return result, nil
		default:
		}

		pageURL := fmt.Sprintf("%s/catalog/%s/?page=%d", baseURL, category.Slug, page)
		slog.Info("scraping page", "category", category.Name, "page", page, "of", totalPages)

		body, err := s.client.Get(ctx, pageURL, category.Slug, s.cookies)
		if err != nil {
			slog.Warn("page failed, stopping early", "page", page, "error", err)
			break
		}

		pageProducts, err := ParseProductsFromHTML(body)
		if err != nil {
			slog.Warn("parse failed, stopping early", "page", page, "error", err)
			break
		}

		for _, p := range pageProducts {
			if _, exists := seen[p.ID]; !exists {
				seen[p.ID] = struct{}{}
				result = append(result, p)
			}
		}
	}

	metrics.ProductsScraped.WithLabelValues(category.Slug).Add(float64(len(result)))
	slog.Info("category done", "name", category.Name, "unique_products", len(result))

	return result, nil
}

const pageSize = 40

type ScrapeResult struct {
	Category Category
	Products []Product
	Err      error
}

func (s *Scraper) ScrapeAll(ctx context.Context, categories []Category, maxPages int) <-chan ScrapeResult {
	results := make(chan ScrapeResult, len(categories))
	go func() {
		defer close(results)
		for _, cat := range categories {
			select {
			case <-ctx.Done():
				return
			default:
			}
			products, err := s.ScrapeCategory(ctx, cat, maxPages)
			results <- ScrapeResult{Category: cat, Products: products, Err: err}
		}
	}()
	return results
}
