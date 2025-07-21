package apperrors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound           = errors.New("запись не найдена")
	ErrAlreadyExists      = errors.New("запись уже существует")
	ErrInvalidCredentials = errors.New("неверное имя пользователя или пароль")
	ErrInvalidToken       = errors.New("неверный или истекший токен")
	ErrUnauthorized       = errors.New("необходима авторизация")
	ErrForbidden          = errors.New("доступ запрещен")
	ErrValidation         = errors.New("ошибка валидации")
	ErrInternal           = errors.New("внутренняя ошибка сервера")
)

var (
	ErrUserNotFound      = fmt.Errorf("пользователь не найден: %w", ErrNotFound)
	ErrUserAlreadyExists = fmt.Errorf("пользователь с таким именем уже существует: %w", ErrAlreadyExists)
	ErrListingNotFound   = fmt.Errorf("объявление не найдено: %w", ErrNotFound)
)

// WrapError оборачивает ошибку с дополнительным контекстом
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFound проверяет, является ли ошибка типа "не найдено"
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists проверяет, является ли ошибка типа "уже существует"
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsUnauthorized проверяет, является ли ошибка типа "необходима авторизация"
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInvalidToken)
}
