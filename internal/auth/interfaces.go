package auth

import (
	"context"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
)

type Repository interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uint64) (*entity.User, error)
}

type UseCase interface {
	Register(ctx context.Context, username, password string) (*entity.UserResponse, error)
	Login(ctx context.Context, username, password string) (string, *entity.UserResponse, error)
	VerifyToken(ctx context.Context, token string) (uint64, error)
}
