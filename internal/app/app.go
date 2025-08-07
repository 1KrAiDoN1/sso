package app

import (
	"context"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage/postgres"
	"sso/pkg/logger"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func NewApp(log *logger.Logger, config config.Config) *App {
	database, err := postgres.NewDatabase(context.Background(), config.DB_config_path)
	if err != nil {
		log.Fatal("NewDatabase", err.Error())
	}
	authService := auth.New(log, database, database, database, config.TokenTTL)

	grpcApp := grpcapp.NewApp(log, authService, config.Port)

	return &App{
		GRPCSrv: grpcApp,
	}
}
