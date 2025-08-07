package grpcapp

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	authgrpc "sso/internal/grpc/auth"
	"sso/internal/services/auth"
	"sso/pkg/logger"
	"syscall"

	"google.golang.org/grpc"
)

type App struct {
	log        *logger.Logger
	gRPCServer *grpc.Server
	port       string
}

func NewApp(log *logger.Logger, authService auth.AuthInterface, port string) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        &logger.Logger{},
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", a.port))
	if err != nil {
		return fmt.Errorf("./internal/app/grpc/app.go: %w", err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("gRPC server is running on port: %s", a.port)
		if err := a.gRPCServer.Serve(l); err != nil {
			serverErr <- fmt.Errorf("gRPC server error: %w", err)
		}
		close(serverErr)
	}()

	// Ожидаем сигналы завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем до получения сигнала или ошибки сервера
	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		log.Print("Shutting down... Received signal: ", sig)
		a.gRPCServer.GracefulStop()
		log.Print("gRPC server stopped")

	}
	return nil

}
