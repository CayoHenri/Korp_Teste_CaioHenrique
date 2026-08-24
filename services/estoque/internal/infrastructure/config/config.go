package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort                   string
	CORSAllowedOrigins         []string
	DatabaseURL                string
	RabbitMQURL                string
	RabbitMQRecoveryMaxRetries int
	RabbitMQRecoveryInterval   time.Duration
	RabbitMQMessageTimeout     time.Duration
	RabbitMQMessageMaxRetries  int
	RabbitMQMessageRetryDelay  time.Duration
}

func Load() (Config, error) {
	loadEnvironmentFile()

	httpPort, err := requiredEnv("ESTOQUE_HTTP_PORT")
	if err != nil {
		return Config{}, err
	}
	corsAllowedOrigins, err := requiredListEnv("ESTOQUE_CORS_ALLOWED_ORIGINS")
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := DatabaseURL()
	if err != nil {
		return Config{}, err
	}
	rabbitMQURL, err := requiredEnv("ESTOQUE_RABBITMQ_URL")
	if err != nil {
		return Config{}, err
	}
	recoveryMaxRetries, recoveryInterval, messageTimeout, messageMaxRetries, messageRetryDelay, err := rabbitMQResilienceConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:                   httpPort,
		CORSAllowedOrigins:         corsAllowedOrigins,
		DatabaseURL:                databaseURL,
		RabbitMQURL:                rabbitMQURL,
		RabbitMQRecoveryMaxRetries: recoveryMaxRetries,
		RabbitMQRecoveryInterval:   recoveryInterval,
		RabbitMQMessageTimeout:     messageTimeout,
		RabbitMQMessageMaxRetries:  messageMaxRetries,
		RabbitMQMessageRetryDelay:  messageRetryDelay,
	}, nil
}

func requiredListEnv(key string) ([]string, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return nil, err
	}

	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%s deve conter ao menos uma origem", key)
	}

	return result, nil
}

func rabbitMQResilienceConfig() (int, time.Duration, time.Duration, int, time.Duration, error) {
	maxRetriesValue, err := requiredEnv("RABBITMQ_RECOVERY_MAX_RETRIES")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	maxRetries, err := strconv.Atoi(maxRetriesValue)
	if err != nil || maxRetries <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("RABBITMQ_RECOVERY_MAX_RETRIES deve ser positivo")
	}
	intervalValue, err := requiredEnv("RABBITMQ_RECOVERY_INTERVAL")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	interval, err := time.ParseDuration(intervalValue)
	if err != nil || interval <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("RABBITMQ_RECOVERY_INTERVAL deve ser uma duracao positiva")
	}
	timeoutValue, err := requiredEnv("RABBITMQ_MESSAGE_TIMEOUT")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil || timeout <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("RABBITMQ_MESSAGE_TIMEOUT deve ser uma duracao positiva")
	}
	messageMaxRetriesValue, err := requiredEnv("RABBITMQ_MESSAGE_MAX_RETRIES")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	messageMaxRetries, err := strconv.Atoi(messageMaxRetriesValue)
	if err != nil || messageMaxRetries <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("RABBITMQ_MESSAGE_MAX_RETRIES deve ser positivo")
	}
	messageRetryDelayValue, err := requiredEnv("RABBITMQ_MESSAGE_RETRY_DELAY")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	messageRetryDelay, err := time.ParseDuration(messageRetryDelayValue)
	if err != nil || messageRetryDelay <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("RABBITMQ_MESSAGE_RETRY_DELAY deve ser uma duracao positiva")
	}
	return maxRetries, interval, timeout, messageMaxRetries, messageRetryDelay, nil
}

func DatabaseURL() (string, error) {
	loadEnvironmentFile()

	return requiredEnv("ESTOQUE_DATABASE_URL")
}

func loadEnvironmentFile() {
	// Variaveis do processo possuem precedencia. A busca ascendente permite
	// executar API, migrations e testes a partir de qualquer pacote do modulo.
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
