package middleware

import (
	"context"
	"strings"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/adapter"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type validator interface {
	Validate() error
}

// ValidationUnaryInterceptor создает унарный интерцептор для валидации входящих запросов
func ValidationUnaryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if isGRPCServiceRequest(info) {
			return handler(ctx, req)
		}
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				log.Error(ctx, "Ошибка валидации запроса",
					zap.String("method", info.FullMethod),
					zap.String("error", err.Error()),
				)
				validationErr := adapter.WrapValidationError(err)
				return nil, adapter.MapError(validationErr)
			}
		}
		return handler(ctx, req)
	}
}

func isGRPCServiceRequest(info *grpc.UnaryServerInfo) bool {
	return strings.HasPrefix(info.FullMethod, "/grpc.") ||
		info.FullMethod == "/grpc.health.v1.Health/Check" ||
		info.FullMethod == wellknown.HealthCheck
}
