// Package model define los tipos compartidos entre microservicios.
package model

import "time"

// OrderStatus representa el estado de un pedido.
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// validTransitions define las transiciones de estado validas para un pedido.
// Cada estado mapea a los estados a los que puede transicionar.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusDelivered},
	StatusDelivered: {},
	StatusCancelled: {},
}

// ValidateTransition verifica si una transicion de estado es valida.
func ValidateTransition(from, to OrderStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// OrderItem representa un articulo dentro de un pedido.
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order representa un pedido en el sistema.
type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CreateOrderArgs son los argumentos para crear un pedido.
type CreateOrderArgs struct {
	UserID string      `json:"user_id"`
	Items  []OrderItem `json:"items"`
}

// ListByUserArgs son los argumentos para listar pedidos de un usuario.
type ListByUserArgs struct {
	UserID string `json:"user_id"`
}

// UpdateStatusArgs son los argumentos para actualizar el estado de un pedido.
type UpdateStatusArgs struct {
	ID     string      `json:"id"`
	Status OrderStatus `json:"status"`
}
