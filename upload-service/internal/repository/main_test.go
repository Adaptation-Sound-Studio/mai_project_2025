package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"upload-service/internal/testutils"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {

	err := godotenv.Load("C:/Users/User/Desktop/Project/mai_project_2025/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	if testing.Short() {
		os.Exit(m.Run())
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf("host=localhost port=5433 user=%s password=%s dbname=anal_db sslmode=disable", user, password)
	testutils.TestDB, err = sql.Open("postgres", connStr)
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
