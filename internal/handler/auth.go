package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/KV2013/gophermart-loyalty/internal/service/auth"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthService interface {
	LoginExists(ctx context.Context, login string) (bool, error)
	Register(ctx context.Context, login, password string) (*model.User, error)
	Authenticate(ctx context.Context, login, password string) (*model.User, error)
}

type AuthHandler struct {
	config      *config.Config
	logger      *zap.Logger
	AuthService AuthService
}

func NewAuthHandler(AuthService AuthService, config *config.Config, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		config:      config,
		logger:      logger,
		AuthService: AuthService,
	}
}

var errBadRequest = errors.New("неверный формат запроса")
var errWriteJSONErrorMsg = errors.New("AuthHandler.APIUserRegister() writeJSONError")
var errLoginExists = errors.New("login уже занят")
var errInternalServerError = errors.New("внутренняя ошибка сервера")
var errInvalidCredentials = errors.New("неверная пара логин/пароль")

// POST /api/user/register — регистрация пользователя;
func (h *AuthHandler) APIUserRegister(res http.ResponseWriter, req *http.Request) {

	var registerRequest RegisterRequest
	writeJSONErrorMsg := "AuthHandler.APIUserRegister() writeJSONError"

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(errWriteJSONErrorMsg.Error(), zap.Error(err))
		}
		return
	}
	err = easyjson.Unmarshal(reqBody, &registerRequest)
	if err != nil {
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	if registerRequest.Login == "" {
		h.logger.Error("AuthHandler.APIUserRegister() не указан login")
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}
	if registerRequest.Password == "" {
		h.logger.Error("AuthHandler.APIUserRegister() не указана password")
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	exists, err := h.AuthService.LoginExists(req.Context(), registerRequest.Login)
	if err != nil {
		h.logger.Error("AuthHandler.APIUserRegister() ошибка проверки существования login", zap.Error(err))
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	if exists {
		h.logger.Error("AuthHandler.APIUserRegister() login уже занят")
		err = writeJSONError(res, errLoginExists, http.StatusConflict)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	user, err := h.AuthService.Register(req.Context(), registerRequest.Login, registerRequest.Password)
	if err != nil {
		h.logger.Error("AuthHandler.APIUserRegister() ошибка регистрации пользователя", zap.Error(err))
		err = writeJSONError(res, errInternalServerError, http.StatusInternalServerError)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	err = h.setAuthCookie(user, res)
	if err != nil {
		h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		err = writeJSONError(res, errInvalidCredentials, http.StatusUnauthorized)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	h.logger.Info("Зарегистрирован новый пользователь", zap.String("login", user.Login))

	res.WriteHeader(http.StatusOK)
}

// POST /api/user/login — аутентификация пользователя;
func (h *AuthHandler) APIUserLogin(res http.ResponseWriter, req *http.Request) {
	const writeJSONErrorMsg = "AuthHandler.APIUserLogin() writeJSONError"

	var loginRequest RegisterRequest

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}
	err = easyjson.Unmarshal(reqBody, &loginRequest)
	if err != nil {
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	if loginRequest.Login == "" || loginRequest.Password == "" {
		err = writeJSONError(res, errBadRequest, http.StatusBadRequest)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	user, err := h.AuthService.Authenticate(req.Context(), loginRequest.Login, loginRequest.Password)
	if err != nil {
		h.logger.Error("AuthHandler.APIUserLogin() ошибка аутентификации", zap.Error(err))
		err = writeJSONError(res, errInvalidCredentials, http.StatusUnauthorized)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	err = h.setAuthCookie(user, res)
	if err != nil {
		h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		err = writeJSONError(res, errInvalidCredentials, http.StatusUnauthorized)
		if err != nil {
			h.logger.Error(writeJSONErrorMsg, zap.Error(err))
		}
		return
	}

	h.logger.Info("Пользователь аутентифицирован", zap.String("login", user.Login))
	res.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) setAuthCookie(user *model.User, res http.ResponseWriter) error {
	tokenString, err := auth.GenerateAccessToken(user.UUID.String(), h.config.JWTSecretKey)
	if err != nil {
		h.logger.Error("AuthHandler.APIUserLogin() ошибка генерации токена", zap.Error(err))
		err = writeJSONError(res, errInternalServerError, http.StatusInternalServerError)
		if err != nil {
			return err
		}

		return nil
	}

	http.SetCookie(res, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(auth.TokenExp),
		HttpOnly: true,
		Path:     "/",
	})

	return nil
}
