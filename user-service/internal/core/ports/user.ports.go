package ports

import "user-service/internal/core/domain"

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByID(id string) (*domain.User, error)
}
