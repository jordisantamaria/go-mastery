package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
)

var (
	ErrNotFound       = errors.New("recurso no encontrado")
	ErrAlreadyExists  = errors.New("el recurso ya existe")
	ErrEmailTaken     = errors.New("el email ya está registrado")
)

// MemoryUserRepository implementa UserRepository usando un mapa en memoria.
// Usa sync.RWMutex para concurrencia segura.
type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*model.User // ID -> User
}

// NewMemoryUserRepository crea una nueva instancia del repositorio de usuarios en memoria.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*model.User),
	}
}

// Create almacena un nuevo usuario. Devuelve error si el email ya existe.
func (r *MemoryUserRepository) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verificar que el email no esté duplicado
	for _, u := range r.users {
		if u.Email == user.Email {
			return ErrEmailTaken
		}
	}

	r.users[user.ID] = user
	return nil
}

// GetByID busca un usuario por su ID.
func (r *MemoryUserRepository) GetByID(_ context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}

	return user, nil
}

// GetByEmail busca un usuario por su email.
func (r *MemoryUserRepository) GetByEmail(_ context.Context, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, ErrNotFound
}

// MemoryTransactionRepository implementa TransactionRepository usando un mapa en memoria.
// Usa sync.RWMutex para concurrencia segura.
type MemoryTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[string]*model.Transaction // ID -> Transaction
}

// NewMemoryTransactionRepository crea una nueva instancia del repositorio de transacciones en memoria.
func NewMemoryTransactionRepository() *MemoryTransactionRepository {
	return &MemoryTransactionRepository{
		transactions: make(map[string]*model.Transaction),
	}
}

// Create almacena una nueva transacción.
func (r *MemoryTransactionRepository) Create(_ context.Context, tx *model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[tx.ID] = tx
	return nil
}

// GetByID busca una transacción por su ID.
func (r *MemoryTransactionRepository) GetByID(_ context.Context, id string) (*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.transactions[id]
	if !ok {
		return nil, ErrNotFound
	}

	return tx, nil
}

// Update actualiza una transacción existente.
func (r *MemoryTransactionRepository) Update(_ context.Context, tx *model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.transactions[tx.ID]; !ok {
		return ErrNotFound
	}

	r.transactions[tx.ID] = tx
	return nil
}

// Delete elimina una transacción por su ID.
func (r *MemoryTransactionRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.transactions[id]; !ok {
		return ErrNotFound
	}

	delete(r.transactions, id)
	return nil
}

// List devuelve las transacciones de un usuario aplicando filtros y paginación.
// Retorna las transacciones de la página actual y el total de resultados.
func (r *MemoryTransactionRepository) List(_ context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Recopilar todas las transacciones del usuario que coincidan con los filtros
	var filtered []model.Transaction
	for _, tx := range r.transactions {
		if tx.UserID != userID {
			continue
		}

		if filter.Type != "" && tx.Type != filter.Type {
			continue
		}

		if filter.Category != "" && tx.Category != filter.Category {
			continue
		}

		if !filter.From.IsZero() && tx.Date.Before(filter.From) {
			continue
		}

		if !filter.To.IsZero() && tx.Date.After(filter.To) {
			continue
		}

		filtered = append(filtered, *tx)
	}

	// Ordenar por fecha descendente (más recientes primero)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date.After(filtered[j].Date)
	})

	total := len(filtered)

	// Aplicar paginación
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}

	start := (page - 1) * limit
	if start >= total {
		return []model.Transaction{}, total, nil
	}

	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}
