package model

import "time"

type APIErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

type ErrUserAlreadyExists struct {
	Login string
}

func (e *ErrUserAlreadyExists) Error() string {
	return "Пользователь с таким логином уже существует login: " + e.Login
}

type ErrUserNotFound struct {
	Login string
}

func (e *ErrUserNotFound) Error() string {
	return "Пользователь не найден login: " + e.Login
}

type ErrUserDeleted struct {
	Login     string
	DeletedAt time.Time
}

func (e *ErrUserDeleted) Error() string {
	return "Пользователь был удалёно login: " + e.Login + " deleted_at: " + e.DeletedAt.String()
}

type ErrOrderOwnedByUser struct {
	Number string
}

func (e *ErrOrderOwnedByUser) Error() string {
	return "заказ уже загружен этим пользователем: " + e.Number
}

type ErrOrderOwnedByOther struct {
	Number string
}

func (e *ErrOrderOwnedByOther) Error() string {
	return "заказ уже загружен другим пользователем: " + e.Number
}

type ErrInsufficientBalance struct{}

func (e *ErrInsufficientBalance) Error() string {
	return "на счету недостаточно средств"
}
