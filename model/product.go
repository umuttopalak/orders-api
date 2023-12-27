package model

type Item struct {
	ItemID    uint64   `json:"item_id"`
	ItemName  string   `json:"item_name"`
	ItemPrice int64    `json:"item_price"`
	Category  Category `json:"category"`
}

type Category struct {
	CategoryID   uint64 `json:"category_id"`
	CategoryName string `json:"category_name"`
}
