// Ejemplo: Servidor HTTP completo con Go 1.22+ routing, middleware, JSON, y graceful shutdown.
//
// Ejecutar: go run ./01-foundations/09-stdlib-deep-dive/examples/http_server.go
//
// Probar con curl:
//   curl http://localhost:8080/api/users
//   curl -X POST http://localhost:8080/api/users -d '{"name":"Alice","email":"alice@example.com"}'
//   curl http://localhost:8080/api/users/42
//
// Para detener el servidor: Ctrl+C (graceful shutdown)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// =============================================================================
// Modelo de datos
// =============================================================================

// User representa un usuario en nuestra API.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// userStore es un almacen en memoria thread-safe para usuarios.
type userStore struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
}

func newUserStore() *userStore {
	return &userStore{
		users:  make(map[int]User),
		nextID: 1,
	}
}

func (s *userStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]User, 0, len(s.users))
	for _, u := range s.users {
		result = append(result, u)
	}
	return result
}

func (s *userStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *userStore) Create(name, email string) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := User{ID: s.nextID, Name: name, Email: email}
	s.users[s.nextID] = u
	s.nextID++
	return u
}

// =============================================================================
// Handlers
// =============================================================================

// listUsers maneja GET /api/users — devuelve todos los usuarios en JSON.
func listUsers(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := store.List()
		writeJSON(w, http.StatusOK, users)
	}
}

// createUser maneja POST /api/users — crea un usuario desde el body JSON.
func createUser(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		// Decodificar JSON del body
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "JSON invalido: " + err.Error(),
			})
			return
		}

		if input.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "el campo 'name' es obligatorio",
			})
			return
		}

		user := store.Create(input.Name, input.Email)
		writeJSON(w, http.StatusCreated, user)
	}
}

// getUser maneja GET /api/users/{id} — devuelve un usuario por ID.
// Usa el path parameter de Go 1.22+ con r.PathValue("id").
func getUser(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path parameter de Go 1.22+
		idStr := r.PathValue("id")
		var id int
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "ID invalido: " + idStr,
			})
			return
		}

		user, ok := store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": fmt.Sprintf("usuario %d no encontrado", id),
			})
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

// writeJSON es un helper para escribir respuestas JSON.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("error escribiendo JSON response", "err", err)
	}
}

// =============================================================================
// Middlewares
// =============================================================================

// loggingMiddleware registra cada request con slog (structured logging).
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Envolver ResponseWriter para capturar el status code
		wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("request completado",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

// statusRecorder envuelve http.ResponseWriter para capturar el status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// recoveryMiddleware captura panics y devuelve 500 en vez de crashear el servidor.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recuperado",
					"error", fmt.Sprintf("%v", err),
					"method", r.Method,
					"path", r.URL.Path,
				)
				http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// =============================================================================
// HTTP Client con timeout
// =============================================================================

// demoHTTPClient muestra como usar http.Client con timeout.
func demoHTTPClient() {
	fmt.Println("\n=== HTTP Client con timeout ===")

	// SIEMPRE configurar timeout — el default no tiene timeout (peligroso)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Tambien se puede usar context para control mas fino
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get", nil)
	if err != nil {
		slog.Error("error creando request", "err", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		// Puede ser timeout, conexion rechazada, etc.
		slog.Warn("request fallo (esperado si no hay red)", "err", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Configurar slog con JSON handler para produccion
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Crear store de usuarios con datos iniciales
	store := newUserStore()
	store.Create("Bob", "bob@example.com")
	store.Create("Charlie", "charlie@example.com")

	// Configurar rutas con Go 1.22+ enhanced routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/users", listUsers(store))
	mux.HandleFunc("POST /api/users", createUser(store))
	mux.HandleFunc("GET /api/users/{id}", getUser(store))

	// Encadenar middlewares: recovery -> logging -> router
	handler := recoveryMiddleware(loggingMiddleware(mux))

	// Configurar servidor con timeouts de produccion
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Demostrar HTTP client (comentado para no bloquear; descomenta para probar)
	// demoHTTPClient()

	// --- Graceful shutdown ---
	// Correr servidor en goroutine
	go func() {
		slog.Info("servidor iniciado", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("error del servidor: %v", err)
		}
	}()

	// Esperar signal de shutdown (Ctrl+C o kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("signal recibido, iniciando shutdown...", "signal", sig.String())

	// Dar 30 segundos para que las conexiones activas terminen
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("error en shutdown", "err", err)
	}

	slog.Info("servidor detenido correctamente")
}
