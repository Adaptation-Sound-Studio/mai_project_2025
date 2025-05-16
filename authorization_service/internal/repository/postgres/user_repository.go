package postgres

import (
	"auth_service/internal/domain/user"
	"database/sql"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user *user.User) error {
	query := `
		INSERT INTO users (name, login, password, is_deleted) 
		VALUES ($1, $2, $3, $4) 
		RETURNING user_id
	`
	return r.DB.QueryRow(query, user.Name, user.Login, user.Pass, user.IsDeleted).Scan(&user.ID)
}

func (r *UserRepository) GetByLogin(login string) (*user.User, error) {
	u := &user.User{}
	query := `SELECT user_id, name, login, password, is_deleted FROM users WHERE login = $1 AND is_deleted = false`
	err := r.DB.QueryRow(query, login).Scan(&u.ID, &u.Name, &u.Login, &u.Pass, &u.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
