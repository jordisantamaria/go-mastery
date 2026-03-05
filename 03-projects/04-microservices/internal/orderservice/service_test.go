package orderservice

import (
	"errors"
	"testing"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// mockUserValidator es un mock del validador de usuarios para tests.
// Permite controlar si la validacion pasa o falla sin depender del UserService real.
type mockUserValidator struct {
	validUsers map[string]bool
}

func newMockValidator(userIDs ...string) *mockUserValidator {
	m := &mockUserValidator{validUsers: make(map[string]bool)}
	for _, id := range userIDs {
		m.validUsers[id] = true
	}
	return m
}

func (m *mockUserValidator) ValidateUser(userID string) error {
	if m.validUsers[userID] {
		return nil
	}
	return ErrUserNotFound
}

// TestCreateOrder verifica que se puede crear un pedido correctamente.
func TestCreateOrder(t *testing.T) {
	validator := newMockValidator("user-1")
	svc := NewService(validator)

	args := model.CreateOrderArgs{
		UserID: "user-1",
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Teclado", Quantity: 1, Price: 49.99},
			{ProductID: "prod-2", Name: "Raton", Quantity: 2, Price: 19.99},
		},
	}

	var order model.Order
	if err := svc.Create(args, &order); err != nil {
		t.Fatalf("error al crear pedido: %v", err)
	}

	if order.ID == "" {
		t.Error("el ID del pedido no deberia estar vacio")
	}
	if order.UserID != "user-1" {
		t.Errorf("UserID esperado 'user-1', obtenido '%s'", order.UserID)
	}
	if order.Status != model.StatusPending {
		t.Errorf("estado esperado 'pending', obtenido '%s'", order.Status)
	}

	expectedTotal := 49.99 + (19.99 * 2)
	if order.Total != expectedTotal {
		t.Errorf("total esperado %.2f, obtenido %.2f", expectedTotal, order.Total)
	}
}

// TestCreateOrderInvalidUser verifica que no se puede crear un pedido para un usuario inexistente.
func TestCreateOrderInvalidUser(t *testing.T) {
	validator := newMockValidator() // Sin usuarios validos
	svc := NewService(validator)

	args := model.CreateOrderArgs{
		UserID: "user-fantasma",
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Teclado", Quantity: 1, Price: 49.99},
		},
	}

	var order model.Order
	err := svc.Create(args, &order)
	if err != ErrUserNotFound {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrUserNotFound, err)
	}
}

// TestCreateOrderNoItems verifica que no se puede crear un pedido sin articulos.
func TestCreateOrderNoItems(t *testing.T) {
	validator := newMockValidator("user-1")
	svc := NewService(validator)

	args := model.CreateOrderArgs{
		UserID: "user-1",
		Items:  []model.OrderItem{},
	}

	var order model.Order
	err := svc.Create(args, &order)
	if err != ErrNoItems {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrNoItems, err)
	}
}

// TestGetOrderByID verifica que se puede obtener un pedido por ID.
func TestGetOrderByID(t *testing.T) {
	validator := newMockValidator("user-1")
	svc := NewService(validator)

	// Crear pedido
	createArgs := model.CreateOrderArgs{
		UserID: "user-1",
		Items:  []model.OrderItem{{ProductID: "p1", Name: "Item", Quantity: 1, Price: 10.0}},
	}
	var created model.Order
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear pedido: %v", err)
	}

	// Buscar por ID
	var found model.Order
	if err := svc.GetByID(model.GetByIDArgs{ID: created.ID}, &found); err != nil {
		t.Fatalf("error al buscar pedido: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("ID esperado '%s', obtenido '%s'", created.ID, found.ID)
	}
}

// TestGetOrderNotFound verifica el error cuando el pedido no existe.
func TestGetOrderNotFound(t *testing.T) {
	validator := newMockValidator()
	svc := NewService(validator)

	var order model.Order
	err := svc.GetByID(model.GetByIDArgs{ID: "no-existe"}, &order)
	if err != ErrOrderNotFound {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrOrderNotFound, err)
	}
}

// TestListByUser verifica que se pueden listar los pedidos de un usuario.
func TestListByUser(t *testing.T) {
	validator := newMockValidator("user-1", "user-2")
	svc := NewService(validator)

	// Crear pedidos para user-1
	for i := 0; i < 3; i++ {
		args := model.CreateOrderArgs{
			UserID: "user-1",
			Items:  []model.OrderItem{{ProductID: "p1", Name: "Item", Quantity: 1, Price: 10.0}},
		}
		var order model.Order
		if err := svc.Create(args, &order); err != nil {
			t.Fatalf("error al crear pedido %d: %v", i, err)
		}
	}

	// Crear pedido para user-2
	args := model.CreateOrderArgs{
		UserID: "user-2",
		Items:  []model.OrderItem{{ProductID: "p2", Name: "Otro", Quantity: 1, Price: 20.0}},
	}
	var order model.Order
	if err := svc.Create(args, &order); err != nil {
		t.Fatalf("error al crear pedido: %v", err)
	}

	// Listar pedidos de user-1
	var orders []model.Order
	if err := svc.ListByUser(model.ListByUserArgs{UserID: "user-1"}, &orders); err != nil {
		t.Fatalf("error al listar pedidos: %v", err)
	}

	if len(orders) != 3 {
		t.Errorf("se esperaban 3 pedidos, se obtuvieron %d", len(orders))
	}
}

// TestUpdateStatus verifica las transiciones de estado de un pedido.
func TestUpdateStatus(t *testing.T) {
	validator := newMockValidator("user-1")
	svc := NewService(validator)

	// Crear pedido (estado: pending)
	createArgs := model.CreateOrderArgs{
		UserID: "user-1",
		Items:  []model.OrderItem{{ProductID: "p1", Name: "Item", Quantity: 1, Price: 10.0}},
	}
	var created model.Order
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear pedido: %v", err)
	}

	// Transicion valida: pending → confirmed
	var updated model.Order
	err := svc.UpdateStatus(model.UpdateStatusArgs{
		ID:     created.ID,
		Status: model.StatusConfirmed,
	}, &updated)
	if err != nil {
		t.Fatalf("error al actualizar estado: %v", err)
	}
	if updated.Status != model.StatusConfirmed {
		t.Errorf("estado esperado 'confirmed', obtenido '%s'", updated.Status)
	}

	// Transicion valida: confirmed → shipped
	err = svc.UpdateStatus(model.UpdateStatusArgs{
		ID:     created.ID,
		Status: model.StatusShipped,
	}, &updated)
	if err != nil {
		t.Fatalf("error al actualizar estado: %v", err)
	}
	if updated.Status != model.StatusShipped {
		t.Errorf("estado esperado 'shipped', obtenido '%s'", updated.Status)
	}
}

// TestUpdateStatusInvalidTransition verifica que no se permiten transiciones invalidas.
func TestUpdateStatusInvalidTransition(t *testing.T) {
	validator := newMockValidator("user-1")
	svc := NewService(validator)

	// Crear pedido (estado: pending)
	createArgs := model.CreateOrderArgs{
		UserID: "user-1",
		Items:  []model.OrderItem{{ProductID: "p1", Name: "Item", Quantity: 1, Price: 10.0}},
	}
	var created model.Order
	if err := svc.Create(createArgs, &created); err != nil {
		t.Fatalf("error al crear pedido: %v", err)
	}

	// Transicion invalida: pending → delivered (sin pasar por confirmed y shipped)
	var updated model.Order
	err := svc.UpdateStatus(model.UpdateStatusArgs{
		ID:     created.ID,
		Status: model.StatusDelivered,
	}, &updated)

	if err == nil {
		t.Fatal("se esperaba un error de transicion invalida")
	}
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("error esperado '%v', obtenido '%v'", ErrInvalidTransition, err)
	}
}
