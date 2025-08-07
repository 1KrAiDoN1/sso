package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DB_config_path string
	GRPCServer     `yaml:"grpc_server"`
}

func SetDBConfig() (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Loading env Config failed, error: %w", err.Error())
		return "", err
	}
	DB_config_path := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	return DB_config_path, nil
}

type GRPCServer struct {
	Port     string        `yaml:"port" env-default:"44044"`
	Timeout  time.Duration `yaml:"timeout" env-default:"4s"`
	TokenTTL time.Duration `yaml:"token_ttl" env-default:"1h"`
}

func MustLoadConfig(server_configPath string) (Config, error) {
	db_configPath, err := SetDBConfig()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(server_configPath)
	if err != nil {
		log.Fatal("Reading Config file failed", "error: %w", err.Error())
		return Config{}, fmt.Errorf("Reading Config file failed %s: %w", server_configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("Parsing YAML failed", "error: %w", err.Error())
		return Config{}, fmt.Errorf("Failed to parse YAML: %w", err)
	}

	return Config{
		DB_config_path: db_configPath,
		GRPCServer: GRPCServer{
			Port:     cfg.Port,
			Timeout:  cfg.Timeout,
			TokenTTL: cfg.TokenTTL,
		},
	}, nil
}
