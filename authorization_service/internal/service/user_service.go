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
