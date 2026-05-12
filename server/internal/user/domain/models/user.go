package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	PasswordHash  string         `json:"-"`
	Role          string         `json:"role"` // 'user' или 'moderator'
	About         sql.NullString `json:"about,omitempty"`
	City          string         `json:"city"`
	Faculty       sql.NullString `json:"faculty,omitempty"`
	Course        sql.NullString `json:"course,omitempty"`
	AvatarImageID sql.NullString `json:"avatar_image_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
