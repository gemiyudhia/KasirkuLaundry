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

type ListFilter struct {
	Status string
	Limit  int
	Offset int
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

func (r *OrderRepository) List(ctx context.Context, f ListFilter) ([]domain.Order, int, error) {
	where := ""
	args := []any{}
	if f.Status != "" {
		where = " WHERE status = ?"
		args = append(args, f.Status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, order_code, customer_name, customer_phone, status, total_amount, notes, created_at, updated_at
	          FROM orders` + where + ` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := []domain.Order{}
	for rows.Next() {
		var (
			o      domain.Order
			status string
			notes  sql.NullString
		)
		if err := rows.Scan(&o.ID, &o.OrderCode, &o.CustomerName, &o.CustomerPhone, &status,
			&o.TotalAmount, &notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		o.Status = domain.OrderStatus(status)
		o.Notes = notes.String
		orders = append(orders, o)
	}
	return orders, total, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, from, to domain.OrderStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = ? WHERE id = ? AND status = ?`, to, id, from)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrInvalidStatusTransition
	}
	return nil
}
