package config

import "testing"

func TestLoadRequiresAllEnvironmentVariables(t *testing.T) {
	t.Setenv("ESTOQUE_HTTP_PORT", "")
	t.Setenv("ESTOQUE_CORS_ALLOWED_ORIGINS", "")
	t.Setenv("ESTOQUE_DATABASE_URL", "")
	t.Setenv("ESTOQUE_RABBITMQ_URL", "")
	t.Setenv("RABBITMQ_RECOVERY_MAX_RETRIES", "")

	if _, err := Load(); err == nil {
		t.Fatal("esperava erro quando as variaveis obrigatorias nao estao definidas")
	}
}

func TestLoadReturnsEnvironmentConfiguration(t *testing.T) {
	t.Setenv("ESTOQUE_HTTP_PORT", "8081")
	t.Setenv("ESTOQUE_CORS_ALLOWED_ORIGINS", "http://localhost:4200")
	t.Setenv("ESTOQUE_DATABASE_URL", "postgres://user:password@localhost:5432/database")
	t.Setenv("ESTOQUE_RABBITMQ_URL", "amqp://user:password@localhost:5672/vhost")
	t.Setenv("RABBITMQ_RECOVERY_MAX_RETRIES", "10")
	t.Setenv("RABBITMQ_RECOVERY_INTERVAL", "1s")
	t.Setenv("RABBITMQ_MESSAGE_TIMEOUT", "5s")
	t.Setenv("RABBITMQ_MESSAGE_MAX_RETRIES", "5")
	t.Setenv("RABBITMQ_MESSAGE_RETRY_DELAY", "1s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	if cfg.HTTPPort != "8081" {
		t.Fatalf("esperava porta 8081, recebeu %s", cfg.HTTPPort)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:4200" {
		t.Fatalf("origens CORS inesperadas: %v", cfg.CORSAllowedOrigins)
	}
}
