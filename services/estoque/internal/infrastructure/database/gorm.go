package database

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connection struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

func Open(databaseURL string) (*Connection, error) {
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("abrir conexao GORM: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("obter conexao SQL: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("testar conexao PostgreSQL: %w", err)
	}

	return &Connection{Gorm: gormDB, SQL: sqlDB}, nil
}

func (connection *Connection) Close() error {
	return connection.SQL.Close()
}
