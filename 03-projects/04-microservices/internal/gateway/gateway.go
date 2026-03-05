// Package gateway implementa el API Gateway REST que traduce peticiones HTTP a llamadas RPC.
// Es el punto de entrada para clientes externos que no hablan RPC directamente.
package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/rpc"
	"strings"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// Gateway es el API Gateway REST que conecta con los servicios RPC.
type Gateway struct {
	userServiceAddr  string
	orderServiceAddr string
	logger           *slog.Logger
}

// NewGateway crea un nuevo Gateway.
func NewGateway(userAddr, orderAddr string, logger *slog.Logger) *Gateway {
	return &Gateway{
		userServiceAddr:  userAddr,
		orderServiceAddr: orderAddr,
		logger:           logger,
	}
}

// RegisterRoutes registra todas las rutas del gateway en un mux.
func (g *Gateway) RegisterRoutes(mux *http.ServeMux) {
	// Rutas de usuarios
	mux.HandleFunc("/api/users", g.handleUsers)
	mux.HandleFunc("/api/users/", g.handleUserByID)

	// Rutas de pedidos
	mux.HandleFunc("/api/orders", g.handleOrders)
	mux.HandleFunc("/api/orders/", g.handleOrderByID)
}

// --- Handlers de usuarios ---

// handleUsers gestiona POST /api/users y GET /api/users
func (g *Gateway) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.createUser(w, r)
	case http.MethodGet:
		g.listUsers(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo no permitido")
	}
}

// handleUserByID gestiona GET/PUT/DELETE /api/users/{id} y GET /api/users/{id}/orders
func (g *Gateway) handleUserByID(w http.ResponseWriter, r *http.Request) {
	// Parsear la ruta: /api/users/{id} o /api/users/{id}/orders
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "ID de usuario requerido")
		return
	}

	userID := parts[0]

	// Si la ruta es /api/users/{id}/orders
	if len(parts) == 2 && parts[1] == "orders" {
		g.listOrdersByUser(w, r, userID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		g.getUserByID(w, r, userID)
	case http.MethodPut:
		g.updateUser(w, r, userID)
	case http.MethodDelete:
		g.deleteUser(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo no permitido")
	}
}

func (g *Gateway) createUser(w http.ResponseWriter, r *http.Request) {
	var args model.CreateUserArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	var user model.User
	if err := g.callUserService("UserService.Create", args, &user); err != nil {
		g.logger.Error("error al crear usuario", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (g *Gateway) getUserByID(w http.ResponseWriter, _ *http.Request, id string) {
	args := model.GetByIDArgs{ID: id}
	var user model.User
	if err := g.callUserService("UserService.GetByID", args, &user); err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		g.logger.Error("error al buscar usuario", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (g *Gateway) listUsers(w http.ResponseWriter, _ *http.Request) {
	args := model.ListArgs{}
	var users []model.User
	if err := g.callUserService("UserService.List", args, &users); err != nil {
		g.logger.Error("error al listar usuarios", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (g *Gateway) updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	args := model.UpdateUserArgs{ID: id, Name: body.Name, Email: body.Email}
	var user model.User
	if err := g.callUserService("UserService.Update", args, &user); err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		g.logger.Error("error al actualizar usuario", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (g *Gateway) deleteUser(w http.ResponseWriter, _ *http.Request, id string) {
	args := model.DeleteArgs{ID: id}
	var deleted bool
	if err := g.callUserService("UserService.Delete", args, &deleted); err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		g.logger.Error("error al eliminar usuario", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"deleted": deleted})
}

// --- Handlers de pedidos ---

// handleOrders gestiona POST /api/orders
func (g *Gateway) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.createOrder(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo no permitido")
	}
}

// handleOrderByID gestiona GET /api/orders/{id} y PUT /api/orders/{id}/status
func (g *Gateway) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "ID de pedido requerido")
		return
	}

	orderID := parts[0]

	// PUT /api/orders/{id}/status
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPut {
		g.updateOrderStatus(w, r, orderID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		g.getOrderByID(w, r, orderID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo no permitido")
	}
}

func (g *Gateway) createOrder(w http.ResponseWriter, r *http.Request) {
	var args model.CreateOrderArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	var order model.Order
	if err := g.callOrderService("OrderService.Create", args, &order); err != nil {
		if strings.Contains(err.Error(), "no existe") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		g.logger.Error("error al crear pedido", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (g *Gateway) getOrderByID(w http.ResponseWriter, _ *http.Request, id string) {
	args := model.GetByIDArgs{ID: id}
	var order model.Order
	if err := g.callOrderService("OrderService.GetByID", args, &order); err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		g.logger.Error("error al buscar pedido", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (g *Gateway) listOrdersByUser(w http.ResponseWriter, _ *http.Request, userID string) {
	args := model.ListByUserArgs{UserID: userID}
	var orders []model.Order
	if err := g.callOrderService("OrderService.ListByUser", args, &orders); err != nil {
		g.logger.Error("error al listar pedidos", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

func (g *Gateway) updateOrderStatus(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Status model.OrderStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	args := model.UpdateStatusArgs{ID: id, Status: body.Status}
	var order model.Order
	if err := g.callOrderService("OrderService.UpdateStatus", args, &order); err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "no permitida") || strings.Contains(err.Error(), "no se puede") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		g.logger.Error("error al actualizar estado", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// --- Helpers RPC ---

// callUserService hace una llamada RPC al UserService.
func (g *Gateway) callUserService(method string, args any, reply any) error {
	client, err := rpc.Dial("tcp", g.userServiceAddr)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call(method, args, reply)
}

// callOrderService hace una llamada RPC al OrderService.
func (g *Gateway) callOrderService(method string, args any, reply any) error {
	client, err := rpc.Dial("tcp", g.orderServiceAddr)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call(method, args, reply)
}

// --- Helpers HTTP ---

// writeJSON escribe una respuesta JSON con el status code indicado.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError escribe una respuesta de error en formato JSON.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
