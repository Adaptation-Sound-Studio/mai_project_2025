package db

import (
	"database/sql"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgresConnection(cfg *config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
