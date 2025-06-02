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
	testDB, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=Rbkkth3920 dbname=auth_db sslmode=disable")
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

func cleanAll(t *testing.T) {
	_, err := testDB.Exec(`DELETE FROM user_role`)
	require.NoError(t, err)
	_, err = testDB.Exec(`DELETE FROM users`)
	require.NoError(t, err)
}

func TestUserRepository_FullLifecycle(t *testing.T) {
	cleanAll(t)
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

func TestUserRepository_UpdateUserRole(t *testing.T) {
	cleanAll(t)
	repo := NewUserRepository(testDB)

	u := &user.User{
		Name:      "Admin User",
		Login:     "adminuser",
		Pass:      "adminpass",
		IsDeleted: false,
	}

	err := repo.Create(u)
	require.NoError(t, err)
	require.NotZero(t, u.ID)

	err = repo.UpdateUserRole(u.ID, "admin")
	require.NoError(t, err)

	role, err := repo.GetRoleByUserID(u.ID)
	require.NoError(t, err)
	require.Equal(t, "admin", role)
}

func TestUserRepository_GetByLogin_NoRows(t *testing.T) {
	repo := NewUserRepository(testDB)
	cleanAll(t)

	user, err := repo.GetByLogin("nonexistent_login")
	require.NoError(t, err)
	require.Nil(t, user)
}

func TestUserRepository_GetRoleByUserID_NoRows(t *testing.T) {
	repo := NewUserRepository(testDB)
	cleanAll(t)

	u := &user.User{
		Name:      "No Role",
		Login:     "noroleuser",
		Pass:      "secret",
		IsDeleted: false,
	}
	err := repo.Create(u)
	require.NoError(t, err)

	role, err := repo.GetRoleByUserID(u.ID)
	require.NoError(t, err)
	require.Empty(t, role)
}

func TestUserRepository_UpdateUserRole_RoleNotFound(t *testing.T) {
	repo := NewUserRepository(testDB)

	err := repo.UpdateUserRole(1, "nonexistent_role")
	require.Error(t, err)
	require.Contains(t, err.Error(), "role nonexistent_role not found")
}
