package testutils

import (
	"database/sql"
	"fmt"
	"testing"
)

var TestDB *sql.DB

func CleanTables(t *testing.T, db *sql.DB) {
	t.Helper()

	// Отключаем ограничения на время удаления
	_, err := db.Exec(`SET session_replication_role = 'replica'`)
	if err != nil {
		t.Fatalf("failed to disable foreign key checks: %v", err)
	}

	tables := []string{
		"fact_listens",
		"songs",
		"artists",
		"albums",
		"genres",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}

	// Возвращаем проверку внешних ключей обратно
	_, err = db.Exec(`SET session_replication_role = 'origin'`)
	if err != nil {
		t.Fatalf("failed to re-enable foreign key checks: %v", err)
	}
}
