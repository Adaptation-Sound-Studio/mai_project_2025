package user

type UserRepository interface {
	Create(user *User) error
	GetByLogin(login string) (*User, error)
	GetRoleByUserID(userID int64) (string, error)
	GetByID(id int64) (*User, error)
	Update(user *User) error
	UpdateUserRole(userID int64, role string) error
}
