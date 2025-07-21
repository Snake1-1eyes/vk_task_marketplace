package grpc

import (
	"context"
	"errors"
	"time"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/auth"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	auth_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler структура обработчика gRPC запросов
type Handler struct {
	auth_pb.UnimplementedAuthServiceServer
	authUC auth.UseCase
	log    *logger.Logger
}

// New создает новый экземпляр Handler
func New(authUC auth.UseCase, log *logger.Logger) *Handler {
	return &Handler{
		authUC: authUC,
		log:    log,
	}
}

// Register обрабатывает запрос на регистрацию пользователя
func (h *Handler) Register(ctx context.Context, req *auth_pb.RegisterRequest) (*auth_pb.RegisterResponse, error) {
	h.log.Info(ctx, "Запрос на регистрацию пользователя", zap.String("username", req.Username))

	userResp, err := h.authUC.Register(ctx, req.Username, req.Password)
	if err != nil {
		h.log.Error(ctx, "Ошибка при регистрации пользователя", zap.Error(err))

		if errors.Is(err, app_errors.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "пользователь с таким именем уже существует")
		}

		return nil, status.Error(codes.Internal, "ошибка при регистрации пользователя")
	}

	h.log.Info(ctx, "Пользователь успешно зарегистрирован", zap.Uint64("user_id", userResp.ID))

	return &auth_pb.RegisterResponse{
		User: &auth_pb.User{
			Id:        userResp.ID,
			Username:  userResp.Username,
			CreatedAt: userResp.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// Login обрабатывает запрос на авторизацию пользователя
func (h *Handler) Login(ctx context.Context, req *auth_pb.LoginRequest) (*auth_pb.LoginResponse, error) {
	h.log.Info(ctx, "Запрос на авторизацию пользователя", zap.String("username", req.Username))

	token, userResp, err := h.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		h.log.Warn(ctx, "Ошибка при авторизации пользователя",
			zap.String("username", req.Username),
			zap.Error(err),
		)

		if errors.Is(err, app_errors.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "неверное имя пользователя или пароль")
		}

		return nil, status.Error(codes.Internal, "ошибка при авторизации пользователя")
	}

	h.log.Info(ctx, "Пользователь успешно авторизован", zap.Uint64("user_id", userResp.ID))

	return &auth_pb.LoginResponse{
		Token: token,
		User: &auth_pb.User{
			Id:        userResp.ID,
			Username:  userResp.Username,
			CreatedAt: userResp.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}
