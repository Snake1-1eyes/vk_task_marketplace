package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryLoggerInterceptor создает унарный gRPC интерцептор для логирования
func UnaryLoggerInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := generateRequestID(ctx)
		ctx = context.WithValue(ctx, logger.RequestID, requestID)
		start := time.Now()
		resp, err := handler(ctx, req)

		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		log.Info(ctx, "gRPC запрос завершен",
			zap.String("method", info.FullMethod),
			zap.String("status", code.String()),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", requestID),
		)

		return resp, err
	}
}

func generateRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get("x-request-id"); len(values) > 0 {
			return values[0]
		}
	}

	return uuid.New().String()
}

// HTTPLoggerMiddleware создает HTTP middleware для внедрения логгера в контекст
func HTTPLoggerMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := getRequestID(r)
			ctx := context.WithValue(r.Context(), logger.Key, log)
			ctx = context.WithValue(ctx, logger.RequestID, requestID)

			w.Header().Set("X-Request-ID", requestID)

			log.Info(ctx, "HTTP запрос получен",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("request_id", requestID),
				zap.String("remote_addr", r.RemoteAddr),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getRequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}
