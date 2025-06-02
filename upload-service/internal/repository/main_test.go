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
	connStr := "host=localhost port=5432 user=upload_user password=upload_pass dbname=upload_db sslmode=disable"
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка открытия соединения с БД: %v", err)
	}

	// ВАЖНО: проверим соединение
	if err := testDB.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Запускаем тесты
	code := m.Run()

	// Закрываем соединение
	if err := testDB.Close(); err != nil {
		log.Printf("Ошибка при закрытии БД: %v", err)
	}

	// Выходим с нужным кодом
	os.Exit(code)
}
