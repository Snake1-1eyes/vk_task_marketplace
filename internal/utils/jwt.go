package utils

import (
	"fmt"
	"time"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims содержит данные, хранимые в JWT токене
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTConfig конфигурация JWT
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// GenerateJWT генерирует JWT токен для пользователя
func GenerateJWT(userID uint64, config JWTConfig) (string, error) {
	expirationTime := time.Now().Add(config.TokenDuration)

	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken проверяет токен и возвращает ID пользователя
func VerifyToken(tokenString string, secretKey string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, app_errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, app_errors.ErrInvalidToken
}
