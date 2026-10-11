package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/domain"
	"github.com/gemiyudhia/KasirkuLaundry/services/order-service/services/order-service/internal/service"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
)

type orderItemRequest struct {
	ServiceName string  `json:"service_name" validate:"required,max=100"`
	UnitPrice   int64   `json:"unit_price" validate:"required,gt=0"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
}

type createOrderRequest struct {
	CustomerName  string             `json:"customer_name" validate:"required,max=100"`
	CustomerPhone string             `json:"customer_phone" validate:"required,min=8,max=20"`
	Notes         string             `json:"notes" validate:"max=500"`
	Items         []orderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderHandler struct {
	svc      *service.OrderService
	validate *validator.Validate
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &OrderHandler{svc: svc, validate: v}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			fields := map[string]string{}
			for _, fe := range verrs {
				ns := fe.Namespace()
				if i := strings.Index(ns, "."); i >= 0 {
					ns = ns[i+1:]
				}
				fields[ns] = fe.Tag()
			}
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":  "validation failed",
				"fields": fields,
			})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	in := service.CreateOrderInput{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		Notes:         req.Notes,
	}
	for _, it := range req.Items {
		in.Items = append(in.Items, service.CreateOrderItemInput{
			ServiceName: it.ServiceName,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
		})
	}

	order, err := h.svc.Create(r.Context(), in)
	if err != nil {
		slog.Error("create order failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	order, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
		return
	}
	if err != nil {
		slog.Error("get order failed", "err", err, "id", id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	status := q.Get("status")
	if status != "" && !domain.OrderStatus(status).IsValid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status filter"})
		return
	}
	page, ok := parsePositiveInt(q.Get("page"), 1)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page"})
		return
	}
	limit, ok := parsePositiveInt(q.Get("limit"), 10)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit"})
		return
	}
	if limit > 100 {
		limit = 100
	}

	orders, total, err := h.svc.List(r.Context(), status, page, limit)
	if err != nil {
		slog.Error("list orders failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": orders,
		"meta": map[string]int{"page": page, "limit": limit, "total": total},
	})
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	var req updateStatusRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	next := domain.OrderStatus(req.Status)
	if !next.IsValid() {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":  "validation failed",
			"fields": map[string]string{"status": "oneof"},
		})
		return
	}

	order, err := h.svc.UpdateStatus(r.Context(), id, next)
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "invalid status transition"})
	case err != nil:
		slog.Error("update status failed", "err", err, "id", id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	default:
		writeJSON(w, http.StatusOK, order)
	}
}

func parsePositiveInt(s string, def int) (int, bool) {
	if s == "" {
		return def, true
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
