package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO orders (order_code, customer_name, customer_phone, status, total_amount, notes)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		o.OrderCode, o.CustomerName, o.CustomerPhone, o.Status, o.TotalAmount,
		sql.NullString{String: o.Notes, Valid: o.Notes != ""},
	)
	if err != nil {
		return err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	o.ID = orderID

	for i := range o.Items {
		it := &o.Items[i]
		res, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, service_name, unit_price, quantity, subtotal)
			 VALUES (?, ?, ?, ?, ?)`,
			orderID, it.ServiceName, it.UnitPrice, it.Quantity, it.Subtotal,
		)
		if err != nil {
			return err
		}
		if it.ID, err = res.LastInsertId(); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	var (
		o      domain.Order
		status string
		notes  sql.NullString
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT id, order_code, customer_name, customer_phone, status, total_amount, notes, created_at, updated_at
		 FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.OrderCode, &o.CustomerName, &o.CustomerPhone, &status,
		&o.TotalAmount, &notes, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	o.Status = domain.OrderStatus(status)
	o.Notes = notes.String

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, service_name, unit_price, quantity, subtotal
		 FROM order_items WHERE order_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	o.Items = []domain.OrderItem{}
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ID, &it.ServiceName, &it.UnitPrice, &it.Quantity, &it.Subtotal); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}
