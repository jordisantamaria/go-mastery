package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/repository"
)

var (
	ErrTransactionNotFound = errors.New("transacción no encontrada")
	ErrUnauthorized        = errors.New("no tienes permiso para acceder a esta transacción")
	ErrValidation          = errors.New("error de validación")
)

// TransactionService contiene la lógica de negocio para transacciones.
type TransactionService struct {
	txRepo repository.TransactionRepository
}

// NewTransactionService crea una nueva instancia con inyección de dependencias.
func NewTransactionService(txRepo repository.TransactionRepository) *TransactionService {
	return &TransactionService{
		txRepo: txRepo,
	}
}

// Create crea una nueva transacción validando los datos de entrada.
func (s *TransactionService) Create(ctx context.Context, userID string, tx *model.Transaction) error {
	// Validar campos obligatorios
	if errs := tx.Validate(); len(errs) > 0 {
		return fmt.Errorf("%w: %v", ErrValidation, errs)
	}

	tx.ID = generateID()
	tx.UserID = userID
	tx.CreatedAt = time.Now()

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("error al crear transacción: %w", err)
	}

	return nil
}

// GetByID obtiene una transacción verificando que pertenezca al usuario.
func (s *TransactionService) GetByID(ctx context.Context, userID, txID string) (*model.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("error al buscar transacción: %w", err)
	}

	// Verificar propiedad: un usuario solo puede ver sus propias transacciones
	if tx.UserID != userID {
		return nil, ErrUnauthorized
	}

	return tx, nil
}

// Update actualiza una transacción verificando propiedad y validación.
func (s *TransactionService) Update(ctx context.Context, userID, txID string, updates *model.Transaction) (*model.Transaction, error) {
	// Obtener la transacción existente (incluye verificación de propiedad)
	existing, err := s.GetByID(ctx, userID, txID)
	if err != nil {
		return nil, err
	}

	// Aplicar actualizaciones manteniendo campos inmutables
	if updates.Type != "" {
		existing.Type = updates.Type
	}
	if updates.Amount > 0 {
		existing.Amount = updates.Amount
	}
	if updates.Category != "" {
		existing.Category = updates.Category
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if !updates.Date.IsZero() {
		existing.Date = updates.Date
	}

	// Validar el resultado final
	if errs := existing.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrValidation, errs)
	}

	if err := s.txRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("error al actualizar transacción: %w", err)
	}

	return existing, nil
}

// Delete elimina una transacción verificando propiedad.
func (s *TransactionService) Delete(ctx context.Context, userID, txID string) error {
	// Verificar que la transacción existe y pertenece al usuario
	if _, err := s.GetByID(ctx, userID, txID); err != nil {
		return err
	}

	if err := s.txRepo.Delete(ctx, txID); err != nil {
		return fmt.Errorf("error al eliminar transacción: %w", err)
	}

	return nil
}

// List devuelve las transacciones del usuario aplicando filtros.
func (s *TransactionService) List(ctx context.Context, userID string, filter model.TransactionFilter) ([]model.Transaction, int, error) {
	// Asegurar valores por defecto para paginación
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	results, total, err := s.txRepo.List(ctx, userID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar transacciones: %w", err)
	}

	return results, total, nil
}
