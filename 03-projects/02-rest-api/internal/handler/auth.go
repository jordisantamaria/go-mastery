package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/model"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/service"
)

// AuthHandler maneja los endpoints de autenticación (registro y login).
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler crea una nueva instancia con inyección de dependencias.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// registerResponse es la respuesta del endpoint de registro.
type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// loginResponse es la respuesta del endpoint de login.
type loginResponse struct {
	Token string           `json:"token"`
	User  registerResponse `json:"user"`
}

// Register maneja POST /api/auth/register.
// Crea un nuevo usuario y devuelve sus datos (sin contraseña).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	// Validar campos
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "el email ya está registrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "error al registrar usuario")
		return
	}

	resp := registerResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login maneja POST /api/auth/login.
// Verifica credenciales y devuelve un token JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	// Validar campos
	if errs := req.Validate(); len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	token, user, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "credenciales inválidas")
			return
		}
		writeError(w, http.StatusInternalServerError, "error al iniciar sesión")
		return
	}

	resp := loginResponse{
		Token: token,
		User: registerResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// --- Funciones auxiliares para respuestas JSON ---

// errorResponse es la estructura estándar para errores de la API.
type errorResponse struct {
	Error string `json:"error"`
}

// validationErrorResponse es la estructura para errores de validación con detalle por campo.
type validationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

// writeError escribe una respuesta JSON de error con el código HTTP indicado.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

// writeValidationErrors escribe una respuesta JSON con errores de validación por campo.
func writeValidationErrors(w http.ResponseWriter, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(validationErrorResponse{
		Error:  "errores de validación",
		Fields: fields,
	})
}
