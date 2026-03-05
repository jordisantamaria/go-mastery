// Punto de entrada del API Gateway.
// Traduce peticiones REST a llamadas RPC hacia los servicios backend.
// Incluye health checks, logging y recovery middleware.
// Gestiona el apagado graceful ante senales SIGINT/SIGTERM.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	gw "github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/internal/gateway"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/health"
	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/pkg/middleware"
)

func main() {
	// Configurar logger estructurado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Leer configuracion del entorno
	addr := os.Getenv("GATEWAY_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	orderServiceAddr := os.Getenv("ORDER_SERVICE_ADDR")
	if orderServiceAddr == "" {
		orderServiceAddr = "localhost:50052"
	}

	// Crear gateway
	gateway := gw.NewGateway(userServiceAddr, orderServiceAddr, logger)

	// Configurar health checks
	checker := health.NewChecker()
	checker.Register("user-service", func() error {
		client, err := rpc.Dial("tcp", userServiceAddr)
		if err != nil {
			return err
		}
		client.Close()
		return nil
	})
	checker.Register("order-service", func() error {
		client, err := rpc.Dial("tcp", orderServiceAddr)
		if err != nil {
			return err
		}
		client.Close()
		return nil
	})

	// Configurar rutas
	mux := http.NewServeMux()
	gateway.RegisterRoutes(mux)
	mux.Handle("/health", checker.Handler())

	// Aplicar middlewares
	var handler http.Handler = mux
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Recovery(logger)(handler)

	// Crear servidor HTTP
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor en goroutine
	go func() {
		logger.Info("API Gateway iniciado",
			"addr", addr,
			"user_service_addr", userServiceAddr,
			"order_service_addr", orderServiceAddr,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error en el servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	// Esperar senal de apagado
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("apagando API Gateway...")

	// Dar tiempo para que las peticiones en vuelo terminen
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error durante el apagado", "error", err)
	}

	logger.Info("API Gateway apagado correctamente")
}
