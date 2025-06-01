package postgres

import (
	"auth_service/internal/domain/user"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open("postgres", "host=localhost port=5432 user=auth_user password=auth_pass dbname=account_service sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err := testDB.Ping(); err != nil {
		panic(fmt.Sprintf("DB not available: %v", err))
	}

	code := m.Run()

	_ = testDB.Close()
	os.Exit(code)
}

func cleanUsers(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func TestUserRepository(t *testing.T) {
	cleanUsers(t)
	repo := NewUserRepository(testDB)

	u := &user.User{
		Name:      "John Doe",
		Login:     "johndoe",
		Pass:      "securepassword",
		IsDeleted: false,
	}
	err := repo.Create(u)
	require.NoError(t, err)
	require.NotZero(t, u.ID)

	found, err := repo.GetByLogin("johndoe")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, u.ID, found.ID)

	byID, err := repo.GetByID(u.ID)
	require.NoError(t, err)
	require.Equal(t, u.Login, byID.Login)

	u.Name = "John Updated"
	u.IsDeleted = true
	err = repo.Update(u)
	require.NoError(t, err)

	deleted, err := repo.GetByID(u.ID)
	require.NoError(t, err)
	require.Nil(t, deleted)
}
