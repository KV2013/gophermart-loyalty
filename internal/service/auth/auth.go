package auth

import (
	"errors"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type UserRepository interface {
	FindByLogin(login string) (*model.User, error)
	Create(user *model.User) error
}

type AuthService struct {
	config         *config.Config
	logger         *zap.Logger
	userRepository *UserRepository
}

func NewAuthService(config *config.Config, logger *zap.Logger) *AuthService {
	return &AuthService{
		config: config,
		logger: logger,
	}
}

func (a *AuthService) LoginExists(login string) bool {
	return false
}

func (a *AuthService) Register(login, password string) (*model.User, error) {
	return &model.User{}, errors.New("foo")
}

func (a *AuthService) Authenticate(login, password string) (*model.User, error) {
	return &model.User{}, errors.New("foo")
}

func (AuthService) checkPassword(password string) error {
}

func (AuthService) hashPassword(password string) string {}
