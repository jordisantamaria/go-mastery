package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"os"
	"testing"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/internal/orderservice"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/internal/userservice"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// setupTestServices inicia los servicios RPC en puertos aleatorios para testing.
// Devuelve las direcciones de los servicios y una funcion de cleanup.
func setupTestServices(t *testing.T) (userAddr, orderAddr string, cleanup func()) {
	t.Helper()

	// Iniciar UserService RPC
	userSvc := userservice.NewService()
	userServer := rpc.NewServer()
	if err := userServer.RegisterName("UserService", userSvc); err != nil {
		t.Fatalf("error al registrar UserService: %v", err)
	}

	userLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error al crear listener para UserService: %v", err)
	}
	go func() {
		for {
			conn, err := userLn.Accept()
			if err != nil {
				return
			}
			go userServer.ServeConn(conn)
		}
	}()

	// Iniciar OrderService RPC con validador que apunta al UserService
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	validator := orderservice.NewRPCUserValidator(userLn.Addr().String(), logger)
	orderSvc := orderservice.NewService(validator)
	orderServer := rpc.NewServer()
	if err := orderServer.RegisterName("OrderService", orderSvc); err != nil {
		t.Fatalf("error al registrar OrderService: %v", err)
	}

	orderLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error al crear listener para OrderService: %v", err)
	}
	go func() {
		for {
			conn, err := orderLn.Accept()
			if err != nil {
				return
			}
			go orderServer.ServeConn(conn)
		}
	}()

	cleanup = func() {
		userLn.Close()
		orderLn.Close()
	}

	return userLn.Addr().String(), orderLn.Addr().String(), cleanup
}

// TestGatewayCreateAndGetUser verifica el flujo completo de crear y obtener un usuario via REST.
func TestGatewayCreateAndGetUser(t *testing.T) {
	userAddr, orderAddr, cleanup := setupTestServices(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := NewGateway(userAddr, orderAddr, logger)

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	// Crear usuario
	body := `{"name":"Ana Garcia","email":"ana@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status esperado %d, obtenido %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
		t.Fatalf("error al decodificar respuesta: %v", err)
	}

	if user.Name != "Ana Garcia" {
		t.Errorf("nombre esperado 'Ana Garcia', obtenido '%s'", user.Name)
	}

	// Obtener usuario por ID
	req = httptest.NewRequest(http.MethodGet, "/api/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status esperado %d, obtenido %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var found model.User
	if err := json.NewDecoder(w.Body).Decode(&found); err != nil {
		t.Fatalf("error al decodificar respuesta: %v", err)
	}

	if found.ID != user.ID {
		t.Errorf("ID esperado '%s', obtenido '%s'", user.ID, found.ID)
	}
}

// TestGatewayListUsers verifica que se pueden listar usuarios via REST.
func TestGatewayListUsers(t *testing.T) {
	userAddr, orderAddr, cleanup := setupTestServices(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := NewGateway(userAddr, orderAddr, logger)

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	// Crear usuarios
	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"name":"User %d","email":"user%d@example.com"}`, i, i)
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("error al crear usuario %d: %s", i, w.Body.String())
		}
	}

	// Listar
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status esperado %d, obtenido %d", http.StatusOK, w.Code)
	}

	var users []model.User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatalf("error al decodificar: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("se esperaban 3 usuarios, se obtuvieron %d", len(users))
	}
}

// TestGatewayCreateOrderFlow verifica el flujo completo: crear usuario, crear pedido, consultar pedido.
func TestGatewayCreateOrderFlow(t *testing.T) {
	userAddr, orderAddr, cleanup := setupTestServices(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := NewGateway(userAddr, orderAddr, logger)

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	// 1. Crear usuario
	userBody := `{"name":"Carlos","email":"carlos@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(userBody))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("error al crear usuario: %s", w.Body.String())
	}

	var user model.User
	json.NewDecoder(w.Body).Decode(&user)

	// 2. Crear pedido
	orderBody := fmt.Sprintf(`{
		"user_id": "%s",
		"items": [
			{"product_id": "p1", "name": "Teclado", "quantity": 1, "price": 49.99},
			{"product_id": "p2", "name": "Raton", "quantity": 2, "price": 19.99}
		]
	}`, user.ID)

	req = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(orderBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("error al crear pedido: %s", w.Body.String())
	}

	var order model.Order
	json.NewDecoder(w.Body).Decode(&order)

	if order.Status != model.StatusPending {
		t.Errorf("estado esperado 'pending', obtenido '%s'", order.Status)
	}

	// 3. Obtener pedido
	req = httptest.NewRequest(http.MethodGet, "/api/orders/"+order.ID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("error al obtener pedido: %s", w.Body.String())
	}

	// 4. Listar pedidos del usuario
	req = httptest.NewRequest(http.MethodGet, "/api/users/"+user.ID+"/orders", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("error al listar pedidos: %s", w.Body.String())
	}

	var orders []model.Order
	json.NewDecoder(w.Body).Decode(&orders)

	if len(orders) != 1 {
		t.Errorf("se esperaba 1 pedido, se obtuvieron %d", len(orders))
	}
}

// TestGatewayUpdateOrderStatus verifica la actualizacion de estado via REST.
func TestGatewayUpdateOrderStatus(t *testing.T) {
	userAddr, orderAddr, cleanup := setupTestServices(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := NewGateway(userAddr, orderAddr, logger)

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	// Crear usuario y pedido
	userBody := `{"name":"Elena","email":"elena@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(userBody))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var user model.User
	json.NewDecoder(w.Body).Decode(&user)

	orderBody := fmt.Sprintf(`{"user_id":"%s","items":[{"product_id":"p1","name":"Item","quantity":1,"price":10}]}`, user.ID)
	req = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(orderBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var order model.Order
	json.NewDecoder(w.Body).Decode(&order)

	// Actualizar estado: pending → confirmed
	statusBody := `{"status":"confirmed"}`
	req = httptest.NewRequest(http.MethodPut, "/api/orders/"+order.ID+"/status", bytes.NewBufferString(statusBody))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status esperado %d, obtenido %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var updated model.Order
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Status != model.StatusConfirmed {
		t.Errorf("estado esperado 'confirmed', obtenido '%s'", updated.Status)
	}
}

// TestGatewayUserNotFound verifica el manejo de error 404 para usuarios.
func TestGatewayUserNotFound(t *testing.T) {
	userAddr, orderAddr, cleanup := setupTestServices(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := NewGateway(userAddr, orderAddr, logger)

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/users/no-existe", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status esperado %d, obtenido %d", http.StatusNotFound, w.Code)
	}
}
