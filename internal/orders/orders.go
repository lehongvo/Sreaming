package orders

import "time"

type Type string

const (
	Buy  Type = "BUY"
	Sell Type = "SELL"
)

type Event struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Symbol    string    `json:"symbol"`
	Type      Type      `json:"type"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}
