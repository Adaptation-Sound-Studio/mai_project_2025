package repository

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	connStr := "host=localhost port=5434 user=postgres password=Rbkkth3920 dbname=upl_db sslmode=disable"
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка открытия соединения с БД: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	code := m.Run()

	if err := testDB.Close(); err != nil {
		log.Printf("Ошибка при закрытии БД: %v", err)
	}

	os.Exit(code)
}
