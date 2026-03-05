// Punto de entrada del servicio de usuarios.
// Inicia un servidor RPC en el puerto configurado y gestiona
// el apagado graceful ante senales SIGINT/SIGTERM.
package main

import (
	"context"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/jordi-nyxidiom/go-mastery/03-projects/04-microservices/internal/userservice"
)

func main() {
	// Configurar logger estructurado
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Leer configuracion del entorno
	addr := os.Getenv("USER_SERVICE_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	// Crear e iniciar el servicio
	svc := userservice.NewService()
	server := rpc.NewServer()
	if err := server.RegisterName("UserService", svc); err != nil {
		logger.Error("error al registrar servicio RPC", "error", err)
		os.Exit(1)
	}

	// Iniciar listener TCP
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("error al iniciar listener", "addr", addr, "error", err)
		os.Exit(1)
	}

	logger.Info("UserService iniciado", "addr", addr)

	// Contexto para apagado graceful
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Aceptar conexiones en goroutine
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// Si el contexto fue cancelado, es un apagado esperado
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
	logger.Info("apagando UserService...")

	// Cerrar listener para dejar de aceptar conexiones
	ln.Close()

	logger.Info("UserService apagado correctamente")
}
