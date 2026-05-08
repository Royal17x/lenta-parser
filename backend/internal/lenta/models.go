package lenta

type Store struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Alias       string `json:"alias"`
	AddressFull string `json:"addressFull"`
	RegionID    int    `json:"regionId"`
}

type Category struct {
	Slug string
	Name string
}

type Product struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	StoreID int     `json:"storeId"`
	Count   int     `json:"count"`
	Prices  Prices  `json:"prices"`
	Display Display `json:"display"`
	Rating  Rating  `json:"rating"`
}

type Prices struct {
	Price        int  `json:"price"`
	PriceRegular int  `json:"priceRegular"`
	IsPromo      bool `json:"isPromoactionPrice"`
}

type Display struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

type Rating struct {
	Rate  float64 `json:"rate"`
	Votes int     `json:"votes"`
}

func (p Prices) PriceRubles() float64 {
	return float64(p.Price) / 100
}

func (p Prices) RegularPriceRubles() float64 {
	return float64(p.PriceRegular) / 100
}

func (p Product) PageURL() string {
	return "https://lenta.com/product/" + p.Slug + "/"
}
