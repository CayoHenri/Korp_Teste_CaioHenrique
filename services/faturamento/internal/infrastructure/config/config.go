package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort       string
	DatabaseURL    string
	EstoqueBaseURL string
	RabbitMQURL    string
	OutboxInterval time.Duration
}

func Load() (Config, error) {
	loadEnvironmentFile()
	port, err := requiredEnv("FATURAMENTO_HTTP_PORT")
	if err != nil {
		return Config{}, err
	}
	databaseURL, err := DatabaseURL()
	if err != nil {
		return Config{}, err
	}
	estoqueBaseURL, err := requiredEnv("FATURAMENTO_ESTOQUE_BASE_URL")
	if err != nil {
		return Config{}, err
	}
	rabbitMQURL, err := requiredEnv("FATURAMENTO_RABBITMQ_URL")
	if err != nil {
		return Config{}, err
	}
	outboxIntervalValue, err := requiredEnv("FATURAMENTO_OUTBOX_INTERVAL")
	if err != nil {
		return Config{}, err
	}
	outboxInterval, err := time.ParseDuration(outboxIntervalValue)
	if err != nil || outboxInterval <= 0 {
		return Config{}, fmt.Errorf("FATURAMENTO_OUTBOX_INTERVAL deve ser uma duracao positiva")
	}
	return Config{
		HTTPPort:       port,
		DatabaseURL:    databaseURL,
		EstoqueBaseURL: estoqueBaseURL,
		RabbitMQURL:    rabbitMQURL,
		OutboxInterval: outboxInterval,
	}, nil
}

func DatabaseURL() (string, error) {
	loadEnvironmentFile()
	return requiredEnv("FATURAMENTO_DATABASE_URL")
}

func MigrationDatabaseURL() (string, error) {
	databaseURL, err := DatabaseURL()
	if err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("URL do banco de dados invalida: %w", err)
	}
	query := parsedURL.Query()
	query.Set("x-migrations-table", "faturamento_schema_migrations")
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func loadEnvironmentFile() {
	directory, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		path := filepath.Join(directory, ".env")
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return
		}
		directory = parent
	}
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("variavel de ambiente obrigatoria ausente: %s", key)
	}
	return value, nil
}
