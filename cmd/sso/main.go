package main

import (
	"sso/internal/app"
	"sso/internal/config"
	"sso/pkg/logger"
)

func main() {
	log := logger.New("main")
	config, err := config.MustLoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatal("MustLoadConfig", err.Error())
	}
	application := app.NewApp(log, config)
	application.GRPCSrv.Run()
}
