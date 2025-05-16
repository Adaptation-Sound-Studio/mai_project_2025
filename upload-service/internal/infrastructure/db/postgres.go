package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgresDB() (*sql.DB, error) {
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "Andreevna54321"
	dbname := "music_db"

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return sql.Open("postgres", dsn)
}
