package middleware

import "net/http"

// CORSConfig contiene la configuración para el middleware CORS.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// DefaultCORSConfig devuelve una configuración CORS permisiva para desarrollo.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}
}

// CORS añade las cabeceras CORS a las respuestas y maneja las peticiones preflight (OPTIONS).
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	origins := joinStrings(config.AllowedOrigins)
	methods := joinStrings(config.AllowedMethods)
	headers := joinStrings(config.AllowedHeaders)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			// Responder directamente a peticiones preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// joinStrings une cadenas con coma como separador.
func joinStrings(s []string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += ", " + s[i]
	}
	return result
}
