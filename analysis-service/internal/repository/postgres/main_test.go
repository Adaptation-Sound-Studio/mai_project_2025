package postgres

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/Adaptation-Sound-Studio/mai_project_2025/analysis/analytics-service/internal/testutils"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(m.Run())
	}
	var err error
	testutils.TestDB, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=analytics_db sslmode=disable")
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
