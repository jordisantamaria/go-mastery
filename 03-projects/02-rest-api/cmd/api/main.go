// Package main es el punto de entrada de la API REST de Finance Tracker.
// Configura las dependencias, registra las rutas y arranca el servidor HTTP
// con soporte para apagado graceful (graceful shutdown).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/handler"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/middleware"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/repository"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/02-rest-api/internal/service"
)

func main() {
	// --- Configuración ---
	port := getEnv("PORT", "8080")
	jwtSecret := []byte(getEnv("JWT_SECRET", "super-secret-key-cambiar-en-produccion"))
	tokenTTL := 24 * time.Hour

	// --- Logger estructurado ---
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// --- Repositorios (in-memory, listos para cambiar a PostgreSQL) ---
	userRepo := repository.NewMemoryUserRepository()
	txRepo := repository.NewMemoryTransactionRepository()

	// --- Servicios (lógica de negocio) ---
	authService := service.NewAuthService(userRepo, jwtSecret, tokenTTL)
	txService := service.NewTransactionService(txRepo)

	// --- Handlers HTTP ---
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)
	txHandler := handler.NewTransactionHandler(txService)

	// --- Router con Go 1.22+ ServeMux mejorado ---
	mux := http.NewServeMux()

	// Rutas públicas
	mux.HandleFunc("GET /api/health", healthHandler.ServeHTTP)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	// Rutas protegidas (requieren JWT)
	authMiddleware := middleware.Auth(jwtSecret)

	mux.Handle("GET /api/transactions", authMiddleware(http.HandlerFunc(txHandler.List)))
	mux.Handle("POST /api/transactions", authMiddleware(http.HandlerFunc(txHandler.Create)))
	mux.Handle("GET /api/transactions/{id}", authMiddleware(http.HandlerFunc(txHandler.GetByID)))
	mux.Handle("PUT /api/transactions/{id}", authMiddleware(http.HandlerFunc(txHandler.Update)))
	mux.Handle("DELETE /api/transactions/{id}", authMiddleware(http.HandlerFunc(txHandler.Delete)))

	// --- Cadena de middleware global: Recovery → CORS → Logging ---
	var finalHandler http.Handler = mux
	finalHandler = middleware.Logging(logger)(finalHandler)
	finalHandler = middleware.CORS(middleware.DefaultCORSConfig())(finalHandler)
	finalHandler = middleware.Recovery(logger)(finalHandler)

	// --- Servidor HTTP ---
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      finalHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- Arranque del servidor en goroutine ---
	go func() {
		logger.Info("servidor iniciado", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error al iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	// --- Graceful shutdown: esperar señal de terminación ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("señal recibida, iniciando apagado graceful", "signal", sig.String())

	// Dar un plazo de 10 segundos para terminar peticiones en curso
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error durante el apagado del servidor", "error", err)
		os.Exit(1)
	}

	logger.Info("servidor apagado correctamente")
}

// getEnv devuelve el valor de una variable de entorno o el valor por defecto.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
