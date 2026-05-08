package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Royal17x/lenta-parser/internal/lenta"
)

func SaveCSV(products []lenta.Product, category string) (string, error) {
	filename := fmt.Sprintf("output_%s_%s.csv", category, timestamp())

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{
		"name", "price_rub", "regular_price_rub", "is_promo", "rating", "votes", "url",
	}); err != nil {
		return "", err
	}

	for _, p := range products {
		row := []string{
			p.Name,
			strconv.FormatFloat(p.Prices.PriceRubles(), 'f', 2, 64),
			strconv.FormatFloat(p.Prices.RegularPriceRubles(), 'f', 2, 64),
			strconv.FormatBool(p.Prices.IsPromo),
			strconv.FormatFloat(p.Rating.Rate, 'f', 1, 64),
			strconv.Itoa(p.Rating.Votes),
			p.PageURL(),
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	return filename, nil
}

func SaveJSON(products []lenta.Product, category string) (string, error) {
	filename := fmt.Sprintf("output_%s_%s.json", category, timestamp())

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(products); err != nil {
		return "", fmt.Errorf("encode json: %w", err)
	}

	return filename, nil
}

func timestamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}
