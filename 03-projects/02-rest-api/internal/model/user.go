package model

import "time"

// User representa un usuario registrado en el sistema.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest representa los datos necesarios para registrar un usuario.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Validate verifica que los campos de registro sean válidos.
func (r *RegisterRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.Email == "" {
		errors["email"] = "es obligatorio"
	}

	if r.Password == "" {
		errors["password"] = "es obligatorio"
	} else if len(r.Password) < 6 {
		errors["password"] = "debe tener al menos 6 caracteres"
	}

	if r.Name == "" {
		errors["name"] = "es obligatorio"
	}

	return errors
}

// LoginRequest representa los datos necesarios para iniciar sesión.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate verifica que los campos de login sean válidos.
func (r *LoginRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.Email == "" {
		errors["email"] = "es obligatorio"
	}

	if r.Password == "" {
		errors["password"] = "es obligatorio"
	}

	return errors
}
