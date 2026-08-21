package config

import (
	"fmt"
	"os"
	"path/filepath"
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
