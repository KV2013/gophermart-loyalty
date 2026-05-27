package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	logger         *zap.Logger
	userRepository UserRepository
}

func NewAuthService(userRepository UserRepository, logger *zap.Logger) *authService {
	return &authService{
		logger:         logger,
		userRepository: userRepository,
	}
}

func (a *authService) LoginExists(ctx context.Context, login string) (bool, error) {
	user, err := a.userRepository.FindByLogin(ctx, login)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

func (a *authService) Register(ctx context.Context, login, password string) (*model.User, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("Register: %w", err)
	}
	return a.userRepository.Create(ctx, login, passwordHash)
}

func (a *authService) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	user, err := a.userRepository.FindByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("Authenticate: %w", err)
	}
	if user == nil {
		return nil, &model.ErrUserNotFound{Login: login}
	}
	if user.DeletedAt != nil {
		return nil, &model.ErrUserDeleted{Login: login, DeletedAt: *user.DeletedAt}
	}

	passwordHash, err := a.userRepository.GetPasswordHash(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("Authenticate: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, &model.ErrUserNotFound{Login: login}
	}
	return user, nil
}

func (a *authService) UserExistsByUUID(ctx context.Context, uuid string) (bool, error) {
	if uuid == "" {
		return false, errors.New("не задан uuid")
	}

	return a.userRepository.UUIDExists(ctx, uuid)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashPassword: %w", err)
	}
	return string(hash), nil
}
