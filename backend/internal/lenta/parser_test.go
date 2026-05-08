package lenta

import (
	"os"
	"testing"
)

func TestParseProductsFromHTML(t *testing.T) {
	data, err := os.ReadFile("testdata/catalog.html")
	if err != nil {
		t.Skipf("testdata/catalog.html not found, skipping: %v", err)
	}

	products, err := ParseProductsFromHTML(data)
	if err != nil {
		t.Fatalf("ParseProductsFromHTML failed: %v", err)
	}

	if len(products) == 0 {
		t.Fatal("expected products, got empty slice")
	}

	t.Logf("parsed %d products", len(products))

	// проверяем первый товар
	first := products[0]
	if first.Name == "" {
		t.Error("product name is empty")
	}
	if first.Slug == "" {
		t.Error("product slug is empty")
	}
	if first.Prices.Price == 0 {
		t.Error("product price is zero")
	}

	t.Logf("first product: %s — %.2f руб — %s", first.Name, first.Prices.PriceRubles(), first.PageURL())
}

// BenchmarkParseProductsFromHTML замеряет скорость парсинга HTML страницы
func BenchmarkParseProductsFromHTML(b *testing.B) {
	data, err := os.ReadFile("testdata/catalog.html")
	if err != nil {
		b.Skipf("testdata/catalog.html not found: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseProductsFromHTML(data)
	}
}
