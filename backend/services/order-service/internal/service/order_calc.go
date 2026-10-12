package service

import (
	"math"

	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/domain"
)

func calculateItems(in []CreateOrderItemInput) ([]domain.OrderItem, int64) {
	items := make([]domain.OrderItem, 0, len(in))
	var total int64

	for _, it := range in {
		qty := math.Round(it.Quantity*100) / 100
		subtotal := int64(math.Round(float64(it.UnitPrice) * qty))
		items = append(items, domain.OrderItem{
			ServiceName: it.ServiceName,
			UnitPrice:   it.UnitPrice,
			Quantity:    qty,
			Subtotal:    subtotal,
		})
		total += subtotal
	}
	return items, total
}
