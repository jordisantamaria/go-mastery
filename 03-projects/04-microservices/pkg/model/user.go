// Package model define los tipos compartidos entre microservicios.
// Estos tipos se usan tanto en las llamadas RPC como en las respuestas REST del gateway.
package model

import "time"

// User representa un usuario en el sistema.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserArgs son los argumentos para crear un usuario.
type CreateUserArgs struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetByIDArgs son los argumentos para buscar por ID (compartido entre servicios).
type GetByIDArgs struct {
	ID string `json:"id"`
}

// ListArgs son los argumentos para listar recursos.
type ListArgs struct{}

// UpdateUserArgs son los argumentos para actualizar un usuario.
type UpdateUserArgs struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// DeleteArgs son los argumentos para eliminar un recurso por ID.
type DeleteArgs struct {
	ID string `json:"id"`
}
