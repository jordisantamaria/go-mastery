// Package userservice implementa el servicio de usuarios via net/rpc.
// Usa un almacenamiento en memoria con sync.RWMutex para concurrencia segura.
// En produccion, el store seria una base de datos como PostgreSQL.
package userservice

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/model"
)

// Errores del servicio de usuarios.
var (
	ErrUserNotFound   = errors.New("usuario no encontrado")
	ErrInvalidName    = errors.New("el nombre es obligatorio")
	ErrInvalidEmail   = errors.New("el email es obligatorio")
	ErrEmailExists    = errors.New("el email ya esta registrado")
)

// Service es la implementacion del servicio de usuarios.
// Se registra como servicio RPC con net/rpc.
type Service struct {
	mu    sync.RWMutex
	users map[string]model.User
}

// NewService crea un nuevo Service de usuarios.
func NewService() *Service {
	return &Service{
		users: make(map[string]model.User),
	}
}

// Create crea un nuevo usuario. Metodo RPC.
func (s *Service) Create(args model.CreateUserArgs, reply *model.User) error {
	if args.Name == "" {
		return ErrInvalidName
	}
	if args.Email == "" {
		return ErrInvalidEmail
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verificar que el email no este duplicado
	for _, u := range s.users {
		if u.Email == args.Email {
			return ErrEmailExists
		}
	}

	user := model.User{
		ID:        generateID(),
		Name:      args.Name,
		Email:     args.Email,
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user
	*reply = user
	return nil
}

// GetByID busca un usuario por ID. Metodo RPC.
func (s *Service) GetByID(args model.GetByIDArgs, reply *model.User) error {
	if args.ID == "" {
		return errors.New("el ID es obligatorio")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[args.ID]
	if !ok {
		return ErrUserNotFound
	}

	*reply = user
	return nil
}

// List devuelve todos los usuarios. Metodo RPC.
func (s *Service) List(args model.ListArgs, reply *[]model.User) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]model.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	*reply = users
	return nil
}

// Update actualiza un usuario existente. Metodo RPC.
func (s *Service) Update(args model.UpdateUserArgs, reply *model.User) error {
	if args.ID == "" {
		return errors.New("el ID es obligatorio")
	}
	if args.Name == "" {
		return ErrInvalidName
	}
	if args.Email == "" {
		return ErrInvalidEmail
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[args.ID]
	if !ok {
		return ErrUserNotFound
	}

	// Verificar email duplicado (excluyendo el usuario actual)
	for _, u := range s.users {
		if u.Email == args.Email && u.ID != args.ID {
			return ErrEmailExists
		}
	}

	user.Name = args.Name
	user.Email = args.Email
	s.users[args.ID] = user

	*reply = user
	return nil
}

// Delete elimina un usuario por ID. Metodo RPC.
func (s *Service) Delete(args model.DeleteArgs, reply *bool) error {
	if args.ID == "" {
		return errors.New("el ID es obligatorio")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[args.ID]; !ok {
		return ErrUserNotFound
	}

	delete(s.users, args.ID)
	*reply = true
	return nil
}

// generateID genera un ID unico usando crypto/rand.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("usr_%x", b)
}
