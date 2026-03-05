package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter envuelve http.ResponseWriter para capturar el código de estado.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captura el código de estado antes de escribir la cabecera.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging registra información de cada petición HTTP: método, ruta, código de estado y duración.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Envolver el ResponseWriter para capturar el código de estado
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.Info("petición HTTP",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
