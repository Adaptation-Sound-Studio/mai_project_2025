package user

type UserRepository interface {
	Create(user *User) error
	GetByLogin(login string) (*User, error)
	GetByID(id int64) (*User, error)
	Update(user *User) error
}
