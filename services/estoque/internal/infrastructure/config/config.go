package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() (Config, error) {
	loadEnvironmentFile()

	httpPort, err := requiredEnv("ESTOQUE_HTTP_PORT")
	if err != nil {
		return Config{}, err
	}

	databaseURL, err := DatabaseURL()
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:    httpPort,
		DatabaseURL: databaseURL,
	}, nil
}

func DatabaseURL() (string, error) {
	loadEnvironmentFile()

	return requiredEnv("ESTOQUE_DATABASE_URL")
}

func loadEnvironmentFile() {
	// Variaveis do processo possuem precedencia. Os caminhos atendem execucao
	// na raiz do repositorio e dentro de services/estoque, respectivamente.
	for _, path := range []string{".env", "../../.env"} {
		_ = godotenv.Load(path)
	}
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("variavel de ambiente obrigatoria ausente: %s", key)
	}

	return value, nil
}
