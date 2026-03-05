package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery intercepta panics en los handlers y devuelve un error 500 en lugar de crashear el servidor.
// Registra el stack trace completo para facilitar la depuración.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					logger.Error("panic recuperado",
						"error", err,
						"stack", string(stack),
						"method", r.Method,
						"path", r.URL.Path,
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"error interno del servidor"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
