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

func (s *Scraper) ScrapeCategory(ctx context.Context, category Category) ([]Product, error) {
	url := fmt.Sprintf("%s/catalog/%s/", baseURL, category.Slug)

	slog.Info("scraping category", "name", category.Name, "url", url)

	body, err := s.client.Get(ctx, url, category.Slug, s.cookies)
	if err != nil {
		metrics.ScrapeErrors.WithLabelValues(category.Slug, "http").Inc()
		return nil, fmt.Errorf("fetch category page: %w", err)
	}

	products, err := ParseProductsFromHTML(body)
	if err != nil {
		metrics.ScrapeErrors.WithLabelValues(category.Slug, "parse").Inc()
		return nil, fmt.Errorf("parse products: %w", err)
	}

	metrics.ProductsScraped.WithLabelValues(category.Slug).Add(float64(len(products)))
	slog.Info("category scraped", "name", category.Name, "products", len(products))

	return products, nil
}

type ScrapeResult struct {
	Category Category
	Products []Product
	Err      error
}

func (s *Scraper) ScrapeAll(ctx context.Context, categories []Category) <-chan ScrapeResult {
	results := make(chan ScrapeResult, len(categories))

	go func() {
		defer close(results)

		for _, cat := range categories {
			select {
			case <-ctx.Done():
				slog.Info("scraping cancelled", "reason", ctx.Err())
				return
			default:
			}

			products, err := s.ScrapeCategory(ctx, cat)
			results <- ScrapeResult{
				Category: cat,
				Products: products,
				Err:      err,
			}
		}
	}()

	return results
}
