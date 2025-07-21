package middleware

import (
	"context"
	"strings"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/auth"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userIDKey string

const (
	UserIDKey userIDKey = "user_id"
)

type authRequirement int

const (
	authOptional authRequirement = iota
	authRequired
	authExempt
)

// methodAuthMap определяет требования к аутентификации для различных методов
var methodAuthMap = map[string]authRequirement{
	"/auth.AuthService/Login":                 authExempt,
	"/auth.AuthService/Register":              authExempt,
	"/listings.ListingsService/CreateListing": authRequired,
}

func getAuthRequirement(method string) authRequirement {
	if req, exists := methodAuthMap[method]; exists {
		return req
	}
	return authOptional
}

// AuthInterceptor создает унарный gRPC интерцептор для авторизации
func AuthInterceptor(authUC auth.UseCase, log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod
		authReq := getAuthRequirement(method)

		if authReq == authExempt {
			return handler(ctx, req)
		}
		token, ok := extractToken(ctx, log, method)

		if !ok {
			if authReq == authRequired {
				log.Warn(ctx, "Отказ в доступе: аутентификация обязательна",
					zap.String("method", method))
				return nil, status.Error(codes.Unauthenticated, "требуется авторизация")
			}
			return handler(ctx, req)
		}

		userID, err := authUC.VerifyToken(ctx, token)
		if err != nil {
			log.Warn(ctx, "Ошибка верификации токена",
				zap.String("method", method),
				zap.Error(err),
			)

			if authReq == authRequired {
				return nil, status.Error(codes.Unauthenticated, "недействительный токен")
			}
			return handler(ctx, req)
		}

		newCtx := context.WithValue(ctx, UserIDKey, userID)
		log.Debug(ctx, "Пользователь авторизован",
			zap.Uint64("user_id", userID),
			zap.String("method", method),
		)

		return handler(newCtx, req)
	}
}

// extractToken извлекает токен из метаданных запроса
func extractToken(ctx context.Context, log *logger.Logger, method string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Warn(ctx, "Метаданные не найдены в запросе", zap.String("method", method))
		return "", false
	}

	values := md.Get("Authorization")
	if len(values) == 0 {
		log.Debug(ctx, "Токен авторизации не предоставлен", zap.String("method", method))
		return "", false
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		log.Warn(ctx, "Неверный формат токена", zap.String("method", method))
		return "", false
	}

	return strings.TrimPrefix(authHeader, "Bearer "), true
}

// GetUserID возвращает ID пользователя из контекста
func GetUserID(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint64)
	return userID, ok
}
