package user_dto

import (
	"database/sql"
	"eventor/internal/platform/types"
	"eventor/internal/user/domain/models"
	"time"
)

type HealthCheckResponse struct {
	Status int `json:"status"`
}

// =====================================================
// REQUEST DTOs (входные данные от клиента)
// =====================================================

// CreateUserRequest данные для регистрации нового пользователя
type CreateUserRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Email         string `json:"email" validate:"required,email,max=255"`
	Password      string `json:"password" validate:"required,min=8,max=128"`
	Role          string `json:"role" validate:"oneof=user moderator"`
	About         string `json:"about,omitempty" validate:"omitempty,max=1000"`
	City          string `json:"city" validate:"required,min=2,max=128"`
	Faculty       string `json:"faculty,omitempty" validate:"omitempty,max=255"`
	Course        string `json:"course,omitempty" validate:"omitempty,max=10"`
	AvatarImageID string `json:"avatar_image_id,omitempty" validate:"omitempty,max=128"`
}

// UpdateUserRequest данные для обновления профиля (все поля опциональны)
type UpdateUserRequest struct {
	Name          *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Password      *string `json:"password,omitempty" validate:"omitempty,min=8,max=128"`
	Role          *string `json:"role,omitempty" validate:"omitempty,oneof=user moderator"`
	About         *string `json:"about,omitempty" validate:"omitempty,max=1000"`
	City          *string `json:"city,omitempty" validate:"omitempty,min=2,max=128"`
	Faculty       *string `json:"faculty,omitempty" validate:"omitempty,max=255"`
	Course        *string `json:"course,omitempty" validate:"omitempty,max=10"`
	AvatarImageID *string `json:"avatar_image_id,omitempty" validate:"omitempty,max=128"`
}

// ListUsersRequest параметры пагинации и фильтрации
type ListUsersRequest struct {
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset,omitempty" validate:"omitempty,min=0"`
	Role   string `json:"role,omitempty" validate:"omitempty,oneof=user moderator"`
	City   string `json:"city,omitempty" validate:"omitempty,max=128"`
}

// =====================================================
// RESPONSE DTOs (выходные данные клиенту)
// =====================================================

// UserResponse публичное представление пользователя
type UserResponse struct {
	ID            types.IdType `json:"id"`
	Name          string       `json:"name"`
	Email         string       `json:"email"`
	Role          string       `json:"role"`
	About         *string      `json:"about,omitempty"`
	City          string       `json:"city"`
	Faculty       *string      `json:"faculty,omitempty"`
	Course        *string      `json:"course,omitempty"`
	AvatarImageID *string      `json:"avatar_image_id,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// UserListResponse ответ со списком пользователей и метаданными
type UserListResponse struct {
	Users   []*UserResponse `json:"users"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
}

// LoginResponse ответ при успешной аутентификации
type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"` // JWT или сессионный токен
}

// =====================================================
// КОНВЕРТЕРЫ: DTO ↔ Domain Model
// =====================================================

// ToDomain конвертирует CreateUserRequest в доменную модель
func (req *CreateUserRequest) ToDomain() *models.User {
	return &models.User{
		Name:          req.Name,
		Email:         req.Email,
		PasswordHash:  "", // заполняется в сервисе
		Role:          req.resolveRole(),
		About:         toNullString(req.About),
		City:          req.City,
		Faculty:       toNullString(req.Faculty),
		Course:        toNullString(req.Course),
		AvatarImageID: toNullString(req.AvatarImageID),
	}
}

// ApplyTo применяет изменения из UpdateUserRequest к доменной модели
func (req *UpdateUserRequest) ApplyTo(user *models.User) {
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.About != nil {
		user.About = toNullString(*req.About)
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.Faculty != nil {
		user.Faculty = toNullString(*req.Faculty)
	}
	if req.Course != nil {
		user.Course = toNullString(*req.Course)
	}
	if req.AvatarImageID != nil {
		user.AvatarImageID = toNullString(*req.AvatarImageID)
	}
	// Password обрабатывается отдельно в сервисе (хеширование)
}

// ToResponse конвертирует доменную модель в UserResponse
func ToResponse(u *models.User) *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		About:         nullStringToPtr(u.About),
		City:          u.City,
		Faculty:       nullStringToPtr(u.Faculty),
		Course:        nullStringToPtr(u.Course),
		AvatarImageID: nullStringToPtr(u.AvatarImageID),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// ToListResponse формирует ответ со списком
func ToListResponse(users []*models.User, total, limit, offset int64) *UserListResponse {
	resp := &UserListResponse{
		Users:   make([]*UserResponse, 0, len(users)),
		Total:   total,
		Limit:   int(limit),
		Offset:  int(offset),
		HasMore: offset+limit < total,
	}
	for _, u := range users {
		resp.Users = append(resp.Users, ToResponse(u))
	}
	return resp
}

// =====================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// =====================================================

func (req *CreateUserRequest) resolveRole() string {
	if req.Role == "moderator" {
		return "moderator"
	}
	return "user"
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
