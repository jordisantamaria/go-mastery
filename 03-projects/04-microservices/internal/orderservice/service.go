// Package orderservice implementa el servicio de pedidos via net/rpc.
// Valida la existencia de usuarios llamando al UserService via RPC,
// demostrando comunicacion inter-servicio.
package orderservice

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/rpc"
	"sync"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// Errores del servicio de pedidos.
var (
	ErrOrderNotFound     = errors.New("pedido no encontrado")
	ErrNoItems           = errors.New("el pedido debe tener al menos un articulo")
	ErrInvalidUserID     = errors.New("el user_id es obligatorio")
	ErrInvalidTransition = errors.New("transicion de estado no permitida")
	ErrUserNotFound      = errors.New("el usuario no existe")
	ErrUserServiceDown   = errors.New("no se pudo conectar al servicio de usuarios")
)

// UserValidator define la interfaz para validar usuarios.
// Permite inyectar un mock en los tests sin depender de una conexion RPC real.
type UserValidator interface {
	ValidateUser(userID string) error
}

// RPCUserValidator valida usuarios conectandose al UserService via RPC.
type RPCUserValidator struct {
	addr   string
	logger *slog.Logger
}

// NewRPCUserValidator crea un validador que se conecta al UserService.
func NewRPCUserValidator(addr string, logger *slog.Logger) *RPCUserValidator {
	return &RPCUserValidator{addr: addr, logger: logger}
}

// ValidateUser verifica que un usuario existe llamando al UserService.
func (v *RPCUserValidator) ValidateUser(userID string) error {
	client, err := rpc.Dial("tcp", v.addr)
	if err != nil {
		v.logger.Error("error al conectar con UserService", "addr", v.addr, "error", err)
		return ErrUserServiceDown
	}
	defer client.Close()

	args := model.GetByIDArgs{ID: userID}
	var user model.User
	if err := client.Call("UserService.GetByID", args, &user); err != nil {
		return ErrUserNotFound
	}

	return nil
}

// Service es la implementacion del servicio de pedidos.
type Service struct {
	mu            sync.RWMutex
	orders        map[string]model.Order
	userValidator UserValidator
}

// NewService crea un nuevo Service de pedidos.
// Recibe un UserValidator para verificar la existencia de usuarios.
func NewService(validator UserValidator) *Service {
	return &Service{
		orders:        make(map[string]model.Order),
		userValidator: validator,
	}
}

// Create crea un nuevo pedido. Valida que el usuario exista antes de crear.
// Metodo RPC.
func (s *Service) Create(args model.CreateOrderArgs, reply *model.Order) error {
	if args.UserID == "" {
		return ErrInvalidUserID
	}
	if len(args.Items) == 0 {
		return ErrNoItems
	}

	// Validar que el usuario existe llamando al UserService
	if err := s.userValidator.ValidateUser(args.UserID); err != nil {
		return err
	}

	// Calcular total
	var total float64
	for _, item := range args.Items {
		total += item.Price * float64(item.Quantity)
	}

	now := time.Now()
	order := model.Order{
		ID:        generateOrderID(),
		UserID:    args.UserID,
		Items:     args.Items,
		Total:     total,
		Status:    model.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.orders[order.ID] = order
	s.mu.Unlock()

	*reply = order
	return nil
}

// GetByID busca un pedido por ID. Metodo RPC.
func (s *Service) GetByID(args model.GetByIDArgs, reply *model.Order) error {
	if args.ID == "" {
		return errors.New("el ID es obligatorio")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[args.ID]
	if !ok {
		return ErrOrderNotFound
	}

	*reply = order
	return nil
}

// ListByUser devuelve todos los pedidos de un usuario. Metodo RPC.
func (s *Service) ListByUser(args model.ListByUserArgs, reply *[]model.Order) error {
	if args.UserID == "" {
		return ErrInvalidUserID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var orders []model.Order
	for _, o := range s.orders {
		if o.UserID == args.UserID {
			orders = append(orders, o)
		}
	}

	if orders == nil {
		orders = []model.Order{}
	}

	*reply = orders
	return nil
}

// UpdateStatus actualiza el estado de un pedido.
// Valida que la transicion de estado sea permitida.
// Metodo RPC.
func (s *Service) UpdateStatus(args model.UpdateStatusArgs, reply *model.Order) error {
	if args.ID == "" {
		return errors.New("el ID es obligatorio")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[args.ID]
	if !ok {
		return ErrOrderNotFound
	}

	// Validar transicion de estado
	if !model.ValidateTransition(order.Status, args.Status) {
		return fmt.Errorf("%w: no se puede pasar de '%s' a '%s'",
			ErrInvalidTransition, order.Status, args.Status)
	}

	order.Status = args.Status
	order.UpdatedAt = time.Now()
	s.orders[args.ID] = order

	*reply = order
	return nil
}

// generateOrderID genera un ID unico para pedidos.
func generateOrderID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ord_%x", b)
}
