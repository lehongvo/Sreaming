package models

import "time"

type WebhookEvent struct {
	ID          string                 `json:"id"`
	Platform    string                 `json:"platform"`
	EventType   string                 `json:"event_type"`
	Payload     map[string]interface{} `json:"payload"`
	ReceivedAt  time.Time              `json:"received_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
}

type EnrichedEvent struct {
	WebhookEvent
	EnrichedData map[string]interface{} `json:"enriched_data"`
	EnrichedAt   time.Time              `json:"enriched_at"`
}

type Order struct {
	ID        string    `json:"id"`
	Platform  string    `json:"platform"`
	UserID    string    `json:"user_id"`
	Items     []Item    `json:"items"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CatalogProduct struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	SKU         string            `json:"sku"`
	Price       float64           `json:"price"`
	Stock       int               `json:"stock"`
	PlatformIDs map[string]string `json:"platform_ids"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
