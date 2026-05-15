package models

import (
	"eventor/internal/platform/types"
	"time"
)

type EventAttendee struct {
	ID       types.IdType `json:"id"`
	UserID   types.IdType `json:"user_id"`
	Name     string       `json:"name"`
	AvatarID string       `json:"avatar_image_id,omitempty"`
	JoinedAt time.Time    `json:"joined_at"`
}
