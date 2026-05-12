package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
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
	passwordHash := hashPassword(password)
	return a.userRepository.Create(ctx, login, passwordHash)
}

func (a *authService) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	passwordHash := hashPassword(password)
	return a.userRepository.FindByCredentials(ctx, login, passwordHash)
}

func (a *authService) UserExistsByUUID(ctx context.Context, uuid string) (bool, error) {
	if uuid == "" {
		return false, errors.New("не задан uuid")
	}

	return a.userRepository.UUIDExists(ctx, uuid)
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}
