package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
)

// --- Tests de UserRepository ---

// TestUserCreate comprueba que se puede crear un usuario correctamente.
func TestUserCreate(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
		CreatedAt:    time.Now(),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	// Verificar que se puede recuperar
	found, err := repo.GetByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("error al buscar usuario: %v", err)
	}

	if found.Email != "test@example.com" {
		t.Errorf("email esperado %q, obtenido %q", "test@example.com", found.Email)
	}
}

// TestUserDuplicateEmail comprueba que no se permite crear dos usuarios con el mismo email.
func TestUserDuplicateEmail(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user1 := &model.User{
		ID:    "user-1",
		Email: "test@example.com",
		Name:  "User 1",
	}
	user2 := &model.User{
		ID:    "user-2",
		Email: "test@example.com",
		Name:  "User 2",
	}

	if err := repo.Create(ctx, user1); err != nil {
		t.Fatalf("error al crear usuario 1: %v", err)
	}

	if err := repo.Create(ctx, user2); err != ErrEmailTaken {
		t.Errorf("se esperaba ErrEmailTaken, obtenido: %v", err)
	}
}

// TestUserGetByEmail comprueba la búsqueda de usuario por email.
func TestUserGetByEmail(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{
		ID:    "user-1",
		Email: "test@example.com",
		Name:  "Test User",
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("error al crear usuario: %v", err)
	}

	found, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("error al buscar por email: %v", err)
	}

	if found.ID != "user-1" {
		t.Errorf("ID esperado %q, obtenido %q", "user-1", found.ID)
	}
}

// TestUserNotFound comprueba que buscar un usuario inexistente devuelve ErrNotFound.
func TestUserNotFound(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "inexistente")
	if err != ErrNotFound {
		t.Errorf("se esperaba ErrNotFound, obtenido: %v", err)
	}

	_, err = repo.GetByEmail(ctx, "no@existe.com")
	if err != ErrNotFound {
		t.Errorf("se esperaba ErrNotFound, obtenido: %v", err)
	}
}

// --- Tests de TransactionRepository ---

// TestTransactionCreate comprueba que se puede crear una transacción.
func TestTransactionCreate(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	tx := &model.Transaction{
		ID:       "tx-1",
		UserID:   "user-1",
		Type:     model.Expense,
		Amount:   50.00,
		Category: "food",
		Date:     time.Now(),
	}

	if err := repo.Create(ctx, tx); err != nil {
		t.Fatalf("error al crear transacción: %v", err)
	}

	found, err := repo.GetByID(ctx, "tx-1")
	if err != nil {
		t.Fatalf("error al buscar transacción: %v", err)
	}

	if found.Amount != 50.00 {
		t.Errorf("amount esperado 50.00, obtenido %f", found.Amount)
	}
}

// TestTransactionUpdate comprueba que se puede actualizar una transacción existente.
func TestTransactionUpdate(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	tx := &model.Transaction{
		ID:       "tx-1",
		UserID:   "user-1",
		Type:     model.Expense,
		Amount:   50.00,
		Category: "food",
		Date:     time.Now(),
	}

	if err := repo.Create(ctx, tx); err != nil {
		t.Fatalf("error al crear transacción: %v", err)
	}

	// Actualizar
	tx.Amount = 75.00
	tx.Category = "restaurant"
	if err := repo.Update(ctx, tx); err != nil {
		t.Fatalf("error al actualizar transacción: %v", err)
	}

	found, err := repo.GetByID(ctx, "tx-1")
	if err != nil {
		t.Fatalf("error al buscar transacción actualizada: %v", err)
	}

	if found.Amount != 75.00 {
		t.Errorf("amount esperado 75.00, obtenido %f", found.Amount)
	}

	if found.Category != "restaurant" {
		t.Errorf("category esperado %q, obtenido %q", "restaurant", found.Category)
	}
}

// TestTransactionDelete comprueba que se puede eliminar una transacción.
func TestTransactionDelete(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	tx := &model.Transaction{
		ID:     "tx-1",
		UserID: "user-1",
		Type:   model.Expense,
		Amount: 50.00,
	}

	if err := repo.Create(ctx, tx); err != nil {
		t.Fatalf("error al crear transacción: %v", err)
	}

	if err := repo.Delete(ctx, "tx-1"); err != nil {
		t.Fatalf("error al eliminar transacción: %v", err)
	}

	_, err := repo.GetByID(ctx, "tx-1")
	if err != ErrNotFound {
		t.Errorf("se esperaba ErrNotFound después de eliminar, obtenido: %v", err)
	}
}

// TestTransactionDeleteNotFound comprueba que eliminar una transacción inexistente devuelve error.
func TestTransactionDeleteNotFound(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "inexistente")
	if err != ErrNotFound {
		t.Errorf("se esperaba ErrNotFound, obtenido: %v", err)
	}
}

// TestTransactionListWithFilters comprueba el filtrado y paginación de transacciones.
func TestTransactionListWithFilters(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	now := time.Now()

	// Crear varias transacciones
	transactions := []*model.Transaction{
		{ID: "tx-1", UserID: "user-1", Type: model.Expense, Amount: 50.00, Category: "food", Date: now.AddDate(0, 0, -1)},
		{ID: "tx-2", UserID: "user-1", Type: model.Income, Amount: 1000.00, Category: "salary", Date: now},
		{ID: "tx-3", UserID: "user-1", Type: model.Expense, Amount: 30.00, Category: "transport", Date: now.AddDate(0, 0, -2)},
		{ID: "tx-4", UserID: "user-1", Type: model.Expense, Amount: 20.00, Category: "food", Date: now.AddDate(0, 0, -3)},
		{ID: "tx-5", UserID: "user-2", Type: model.Expense, Amount: 100.00, Category: "food", Date: now}, // Otro usuario
	}

	for _, tx := range transactions {
		if err := repo.Create(ctx, tx); err != nil {
			t.Fatalf("error al crear transacción %s: %v", tx.ID, err)
		}
	}

	// Listar todas las transacciones del usuario 1
	results, total, err := repo.List(ctx, "user-1", model.TransactionFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("error al listar transacciones: %v", err)
	}
	if total != 4 {
		t.Errorf("total esperado 4, obtenido %d", total)
	}
	if len(results) != 4 {
		t.Errorf("resultados esperados 4, obtenido %d", len(results))
	}

	// Filtrar por tipo
	results, total, err = repo.List(ctx, "user-1", model.TransactionFilter{Type: model.Expense, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("error al listar con filtro de tipo: %v", err)
	}
	if total != 3 {
		t.Errorf("total esperado 3 gastos, obtenido %d", total)
	}

	// Filtrar por categoría
	results, total, err = repo.List(ctx, "user-1", model.TransactionFilter{Category: "food", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("error al listar con filtro de categoría: %v", err)
	}
	if total != 2 {
		t.Errorf("total esperado 2 de food, obtenido %d", total)
	}

	// Paginación
	results, total, err = repo.List(ctx, "user-1", model.TransactionFilter{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("error al listar con paginación: %v", err)
	}
	if total != 4 {
		t.Errorf("total esperado 4, obtenido %d", total)
	}
	if len(results) != 2 {
		t.Errorf("resultados en página esperados 2, obtenido %d", len(results))
	}

	// Página fuera de rango
	results, _, err = repo.List(ctx, "user-1", model.TransactionFilter{Page: 100, Limit: 10})
	if err != nil {
		t.Fatalf("error al listar página fuera de rango: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("resultados esperados 0 en página fuera de rango, obtenido %d", len(results))
	}
}

// TestTransactionListByDateRange comprueba el filtrado por rango de fechas.
func TestTransactionListByDateRange(t *testing.T) {
	repo := NewMemoryTransactionRepository()
	ctx := context.Background()

	now := time.Now()

	transactions := []*model.Transaction{
		{ID: "tx-1", UserID: "user-1", Type: model.Expense, Amount: 10, Category: "food", Date: now.AddDate(0, -2, 0)},
		{ID: "tx-2", UserID: "user-1", Type: model.Expense, Amount: 20, Category: "food", Date: now.AddDate(0, -1, 0)},
		{ID: "tx-3", UserID: "user-1", Type: model.Expense, Amount: 30, Category: "food", Date: now},
	}

	for _, tx := range transactions {
		if err := repo.Create(ctx, tx); err != nil {
			t.Fatalf("error al crear transacción: %v", err)
		}
	}

	// Filtrar por rango: solo las del último mes
	from := now.AddDate(0, -1, -1)
	to := now.AddDate(0, 0, 1)

	results, total, err := repo.List(ctx, "user-1", model.TransactionFilter{
		From:  from,
		To:    to,
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("error al listar por rango de fechas: %v", err)
	}

	if total != 2 {
		t.Errorf("total esperado 2 en rango, obtenido %d", total)
	}
	_ = results
}
