package service

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/domain"
	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/repository"
)

type CreateOrderItemInput struct {
	ServiceName string
	UnitPrice   int64
	Quantity    float64
}

type CreateOrderInput struct {
	CustomerName  string
	CustomerPhone string
	Notes         string
	Items         []CreateOrderItemInput
}

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	items := make([]domain.OrderItem, 0, len(in.Items))
	var total int64

	for _, it := range in.Items {
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

	order := &domain.Order{
		OrderCode:     generateOrderCode(),
		CustomerName:  in.CustomerName,
		CustomerPhone: in.CustomerPhone,
		Notes:         in.Notes,
		Status:        domain.StatusReceived,
		TotalAmount:   total,
		Items:         items,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func generateOrderCode() string {
	return fmt.Sprintf("LDY-%s-%04d", time.Now().Format("20060102"), rand.IntN(10000))
}
