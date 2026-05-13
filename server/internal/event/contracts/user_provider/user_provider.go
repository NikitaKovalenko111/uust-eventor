package user_provider

import (
	"context"
	"eventor/internal/user/domain/models"
)

type UserProvider interface {
	GetByID(ctx context.Context, id uint64) (*models.User, error)
}
