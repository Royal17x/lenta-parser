package lenta

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var ngStateRegex = regexp.MustCompile(
	`<script id="ng-state" type="application/json">([\s\S]*?)</script>`,
)

type ngState struct {
	CatalogItems *catalogItemsResponse `json:"/catalog/items"`
}

type catalogItemsResponse struct {
	Items []Product `json:"items"`
	Total int       `json:"total"`
}

func ParseProductsFromHTML(htmlBody []byte) ([]Product, error) {
	matches := ngStateRegex.FindSubmatch(htmlBody)
	if len(matches) < 2 {
		return nil, fmt.Errorf("ng-state script tag not found in HTML")
	}

	var state ngState
	if err := json.Unmarshal(matches[1], &state); err != nil {
		return nil, fmt.Errorf("unmarshal ng-state: %w", err)
	}

	if state.CatalogItems == nil {
		return nil, fmt.Errorf("/catalog/items key not found in ng-state")
	}

	if len(state.CatalogItems.Items) == 0 {
		return nil, fmt.Errorf("items array is empty (total=%d)", state.CatalogItems.Total)
	}

	return state.CatalogItems.Items, nil
}

func ParseTotalFromHTML(htmlBody []byte) (int, error) {
	matches := ngStateRegex.FindSubmatch(htmlBody)
	if len(matches) < 2 {
		return 0, fmt.Errorf("ng-state not found")
	}
	var state ngState
	if err := json.Unmarshal(matches[1], &state); err != nil {
		return 0, err
	}
	if state.CatalogItems == nil {
		return 0, nil
	}
	return state.CatalogItems.Total, nil
}
