package user_provider

import (
	"context"
	"eventor/internal/platform/types"
	"eventor/internal/user/domain/models"
)

type UserProvider interface {
	GetByID(ctx context.Context, id types.IdType) (*models.User, error)
}
