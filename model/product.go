package model

type Product struct {
	ProductID    uint64   `json:"product_id"`
	ProductName  string   `json:"product_name"`
	ProductPrice int64    `json:"product_price"`
	Category     Category `json:"category"`
}
