package config

import "testing"

func TestLoadRequiresAllEnvironmentVariables(t *testing.T) {
	t.Setenv("ESTOQUE_HTTP_PORT", "")
	t.Setenv("ESTOQUE_DATABASE_URL", "")
	t.Setenv("ESTOQUE_RABBITMQ_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("esperava erro quando as variaveis obrigatorias nao estao definidas")
	}
}

func TestLoadReturnsEnvironmentConfiguration(t *testing.T) {
	t.Setenv("ESTOQUE_HTTP_PORT", "8081")
	t.Setenv("ESTOQUE_DATABASE_URL", "postgres://user:password@localhost:5432/database")
	t.Setenv("ESTOQUE_RABBITMQ_URL", "amqp://user:password@localhost:5672/vhost")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}

	if cfg.HTTPPort != "8081" {
		t.Fatalf("esperava porta 8081, recebeu %s", cfg.HTTPPort)
	}
}
