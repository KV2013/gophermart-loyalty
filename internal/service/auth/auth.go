package auth

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type UserRepository interface {
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindByCredentials(ctx context.Context, login string, passwordhash string) (*model.User, error)
	Create(ctx context.Context, login string, passwordhash string) (*model.User, error)
}

type AuthService struct {
	logger         *zap.Logger
	userRepository UserRepository
}

func NewAuthService(userRepository UserRepository, logger *zap.Logger) *AuthService {
	return &AuthService{
		logger:         logger,
		userRepository: userRepository,
	}
}

func (a *AuthService) LoginExists(ctx context.Context, login string) (bool, error) {

	user, err := a.userRepository.FindByLogin(ctx, login)
	if err != nil {
		return false, err
	}

	exists := user != nil

	return exists, nil
}

func (a *AuthService) Register(ctx context.Context, login, password string) (*model.User, error) {
	passwordHash := a.hashPassword(password)
	user, err := a.userRepository.Create(ctx, login, passwordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthService) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	passwordHash := a.hashPassword(password)

	user, err := a.userRepository.FindByCredentials(ctx, login, passwordHash)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (AuthService) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}
