package user_provider

import (
	"context"
	"database/sql"
	"eventor/internal/user/domain/models"
	user_dto "eventor/internal/user/transport/http/dto/user"
)

type UserProvider interface {
	Create(ctx context.Context, req *user_dto.CreateUserRequest, tx *sql.Tx) (*models.User, error)
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
