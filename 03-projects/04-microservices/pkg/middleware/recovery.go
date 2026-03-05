package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery es un middleware que recupera panics y devuelve un error 500.
// Registra el stack trace para debugging sin exponer detalles al cliente.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recuperado",
						"error", err,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
					)
					http.Error(w, `{"error":"error interno del servidor"}`, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
