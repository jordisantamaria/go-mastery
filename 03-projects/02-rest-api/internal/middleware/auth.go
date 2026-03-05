package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/pkg/jwt"
)

// contextKey es un tipo privado para evitar colisiones en las claves del contexto.
type contextKey string

const (
	// UserIDKey es la clave del contexto donde se almacena el ID del usuario autenticado.
	UserIDKey contextKey = "user_id"
	// UserEmailKey es la clave del contexto donde se almacena el email del usuario autenticado.
	UserEmailKey contextKey = "user_email"
)

// Auth verifica el token JWT en la cabecera Authorization.
// Si el token es válido, añade user_id y user_email al contexto de la petición.
// Formato esperado: "Authorization: Bearer <token>"
func Auth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"se requiere autenticación"}`))
				return
			}

			// Extraer el token del formato "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"formato de autorización inválido, use: Bearer <token>"}`))
				return
			}

			token := parts[1]

			// Verificar el token
			claims, err := jwt.Verify(token, jwtSecret)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"token inválido o expirado"}`))
				return
			}

			// Añadir datos del usuario al contexto
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extrae el ID del usuario del contexto de la petición.
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// GetUserEmail extrae el email del usuario del contexto de la petición.
func GetUserEmail(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailKey).(string)
	return email
}
