package model

import "time"

// TransactionType representa el tipo de transacción: ingreso o gasto.
type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

// Transaction representa una transacción financiera del usuario.
type Transaction struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TransactionFilter contiene los filtros opcionales para listar transacciones.
type TransactionFilter struct {
	Type     TransactionType
	Category string
	From     time.Time
	To       time.Time
	Page     int
	Limit    int
}

// Validate verifica que los campos obligatorios de la transacción sean válidos.
func (t *Transaction) Validate() map[string]string {
	errors := make(map[string]string)

	if t.Type != Income && t.Type != Expense {
		errors["type"] = "debe ser 'income' o 'expense'"
	}

	if t.Amount <= 0 {
		errors["amount"] = "debe ser mayor que 0"
	}

	if t.Category == "" {
		errors["category"] = "es obligatorio"
	}

	if t.Date.IsZero() {
		errors["date"] = "es obligatorio"
	}

	return errors
}
