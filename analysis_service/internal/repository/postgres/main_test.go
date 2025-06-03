package postgres

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	flag.Parse()
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
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	if err := testutils.TestDB.Ping(); err != nil {
		log.Fatalf("Test DB is not responding: %v", err)
	}

	code := m.Run()

	testutils.TestDB.Close()
	os.Exit(code)
}
