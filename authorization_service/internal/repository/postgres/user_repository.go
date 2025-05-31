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

func (r *UserRepository) GetRoleByUserID(userID int64) (string, error) {
	var role string
	query := `
		SELECT r.role
		FROM roles r
		INNER JOIN user_role ur ON ur.role_id = r.role_id
		WHERE ur.user_id = $1
		LIMIT 1
	`
	err := r.DB.QueryRow(query, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return role, nil
}

func (r *UserRepository) GetByID(id int64) (*user.User, error) {
	u := &user.User{}
	query := `SELECT user_id, name, login, password, is_deleted FROM users WHERE user_id = $1 AND is_deleted = false`
	err := r.DB.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.Login, &u.Pass, &u.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Update(u *user.User) error {
	query := `
		UPDATE users 
		SET name = $1, login = $2, password = $3, is_deleted = $4 
		WHERE user_id = $5
	`
	_, err := r.DB.Exec(query, u.Name, u.Login, u.Pass, u.IsDeleted, u.ID)
	return err
}
