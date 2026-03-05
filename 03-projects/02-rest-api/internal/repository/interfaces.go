package repository

import (
	"context"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
)

// UserRepository define las operaciones de persistencia para usuarios.
// Cualquier implementación (memoria, PostgreSQL, etc.) debe cumplir esta interfaz.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

// TransactionRepository define las operaciones de persistencia para transacciones.
// Cualquier implementación (memoria, PostgreSQL, etc.) debe cumplir esta interfaz.
type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id string) (*model.Transaction, error)
	Update(ctx context.Context, tx *model.Transaction) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, int, error)
}
