package usecase

import (
	"context"
	"time"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/auth"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/utils"
	"go.uber.org/zap"
)

// UseCase реализует интерфейс auth.UseCase
type UseCase struct {
	repo      auth.Repository
	jwtConfig utils.JWTConfig
	log       *logger.Logger
}

// New создает новый экземпляр UseCase
func New(repo auth.Repository, jwtConfig utils.JWTConfig, log *logger.Logger) *UseCase {
	return &UseCase{
		repo:      repo,
		jwtConfig: jwtConfig,
		log:       log,
	}
}

// Register регистрирует нового пользователя
func (uc *UseCase) Register(ctx context.Context, username, password string) (*entity.UserResponse, error) {
	_, err := uc.repo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, app_errors.ErrUserAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		uc.log.Error(ctx, "Ошибка хеширования пароля", zap.Error(err))
		return nil, app_errors.WrapError(err, "ошибка при регистрации пользователя")
	}

	user := &entity.User{
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	createdUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		uc.log.Error(ctx, "Ошибка создания пользователя", zap.Error(err))
		return nil, app_errors.WrapError(err, "ошибка при регистрации пользователя")
	}

	return createdUser.ToResponse(), nil
}

// Login авторизует пользователя и возвращает токен
func (uc *UseCase) Login(ctx context.Context, username, password string) (string, *entity.UserResponse, error) {
	user, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		uc.log.Error(ctx, "Пользователь не найден", zap.String("username", username), zap.Error(err))
		return "", nil, app_errors.ErrInvalidCredentials
	}

	err = utils.CheckPasswordHash(password, user.Password)
	if err != nil {
		uc.log.Warn(ctx, "Неверный пароль", zap.String("username", username))
		return "", nil, app_errors.ErrInvalidCredentials
	}

	token, err := utils.GenerateJWT(user.ID, uc.jwtConfig)
	if err != nil {
		uc.log.Error(ctx, "Ошибка генерации токена", zap.Error(err))
		return "", nil, app_errors.WrapError(err, "ошибка авторизации")
	}

	return token, user.ToResponse(), nil
}

func (uc *UseCase) VerifyToken(ctx context.Context, tokenString string) (uint64, error) {
	userID, err := utils.VerifyToken(tokenString, uc.jwtConfig.SecretKey)
	if err != nil {
		return 0, err
	}

	_, err = uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return 0, app_errors.ErrInvalidToken
	}

	return userID, nil
}
