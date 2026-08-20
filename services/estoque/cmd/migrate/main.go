package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/infrastructure/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("uso: go run ./cmd/migrate <up|down|version|force> [versao]")
	}

	databaseURL, err := config.DatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	migrator, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Fatalf("inicializar migrations: %v", err)
	}
	defer func() {
		_, _ = migrator.Close()
	}()

	if err := execute(migrator, os.Args[1], os.Args[2:]); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("nenhuma migration pendente")
			return
		}

		log.Fatal(err)
	}
}

func execute(migrator *migrate.Migrate, command string, args []string) error {
	switch command {
	case "up":
		return migrator.Up()
	case "down":
		return migrator.Steps(-1)
	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			return err
		}
		fmt.Printf("versao=%d dirty=%t\n", version, dirty)
		return nil
	case "force":
		if len(args) != 1 {
			return errors.New("uso: go run ./cmd/migrate force <versao>")
		}
		version, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("versao invalida: %w", err)
		}
		return migrator.Force(version)
	default:
		return fmt.Errorf("comando de migration desconhecido: %s", command)
	}
}
