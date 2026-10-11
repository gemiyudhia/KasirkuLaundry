package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	StatusReceived   OrderStatus = "received"
	StatusProcessing OrderStatus = "processing"
	StatusReady      OrderStatus = "ready"
	StatusCompleted  OrderStatus = "completed"
)

var ErrOrderNotFound = errors.New("order not found")

type Order struct {
	ID            int64       `json:"id"`
	OrderCode     string      `json:"order_code"`
	CustomerName  string      `json:"customer_name"`
	CustomerPhone string      `json:"customer_phone"`
	Status        OrderStatus `json:"status"`
	TotalAmount   int64       `json:"total_amount"`
	Notes         string      `json:"notes,omitempty"`
	Items         []OrderItem `json:"items"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID          int64   `json:"id"`
	ServiceName string  `json:"service_name"`
	UnitPrice   int64   `json:"unit_price"`
	Quantity    float64 `json:"quantity"`
	Subtotal    int64   `json:"subtotal"`
}