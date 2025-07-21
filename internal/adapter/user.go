package adapter

import (
	"time"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
	auth_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/auth"
)

// MapUserToProto преобразует внутреннюю модель пользователя в proto-объект
func MapUserToProto(user *entity.UserResponse) *auth_pb.User {
	return &auth_pb.User{
		Id:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
