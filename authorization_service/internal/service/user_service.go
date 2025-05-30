package service

import "auth_service/internal/domain/user"

type UserService struct {
	Repo user.UserRepository
}

func NewUserService(repo user.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) CreateUser(user *user.User) error {
	return s.Repo.Create(user)
}

func (s *UserService) GetUserByID(id int64) (*user.User, error) {
	return s.Repo.GetByID(id)
}

func (s *UserService) UpdateUser(user *user.User) error {
	return s.Repo.Update(user)
}
