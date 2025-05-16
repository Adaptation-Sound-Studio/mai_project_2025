package user

type UserRepository interface {
	Create(user *User) error
	GetByLogin(login string) (*User, error)
}
