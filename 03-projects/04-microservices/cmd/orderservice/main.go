// Punto de entrada del servicio de pedidos.
// Inicia un servidor RPC y se conecta al UserService para validar usuarios.
// Gestiona el apagado graceful ante senales SIGINT/SIGTERM.
package main

import (
	"context"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/internal/orderservice"
)

func main() {
	// Configurar logger estructurado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Leer configuracion del entorno
	addr := os.Getenv("ORDER_SERVICE_ADDR")
	if addr == "" {
		addr = ":50052"
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	// Crear validador de usuarios que conecta al UserService via RPC
	validator := orderservice.NewRPCUserValidator(userServiceAddr, logger)

	// Crear e iniciar el servicio
	svc := orderservice.NewService(validator)
	server := rpc.NewServer()
	if err := server.RegisterName("OrderService", svc); err != nil {
		logger.Error("error al registrar servicio RPC", "error", err)
		os.Exit(1)
	}

	// Iniciar listener TCP
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("error al iniciar listener", "addr", addr, "error", err)
		os.Exit(1)
	}

	logger.Info("OrderService iniciado",
		"addr", addr,
		"user_service_addr", userServiceAddr,
	)

	// Contexto para apagado graceful
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Aceptar conexiones en goroutine
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Error("error al aceptar conexion", "error", err)
					continue
				}
			}
			go server.ServeConn(conn)
		}
	}()

	// Esperar senal de apagado
	<-ctx.Done()
	logger.Info("apagando OrderService...")

	ln.Close()

	logger.Info("OrderService apagado correctamente")
}
