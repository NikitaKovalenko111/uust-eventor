package user_dto

import (
	"database/sql"
	"eventor/internal/platform/types"
	"eventor/internal/user/domain/models"
	"time"
)

// HealthCheckResponse
// @Description Health check response
type HealthCheckResponse struct {
	// HTTP status code indicating service health
	// @example 200
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

// UpdateUserRequest
// @Description Request body for updating user profile. All fields are optional; null means "do not change".
type UpdateUserRequest struct {
	// User's display name (2-255 characters)
	// @example "Ivan Ivanov"
	// @MinLength(2) @MaxLength(255)
	Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	// User's email address (must be valid and unique)
	// @example "user@example.com"
	// @Format email
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	// New password (8-128 characters, only for admin flows)
	// @example "NewSecurePass123!"
	// @MinLength(8) @MaxLength(128)
	Password *string `json:"password,omitempty" validate:"omitempty,min=8,max=128"`
	// User role: "user" or "moderator" (admin-only field)
	// @enum user moderator
	// @example "user"
	Role *string `json:"role,omitempty" validate:"omitempty,oneof=user moderator"`
	// Short bio or about me section (max 1000 characters)
	// @example "Software engineer passionate about Go and distributed systems"
	// @MaxLength(1000)
	About *string `json:"about,omitempty" validate:"omitempty,max=1000"`
	// User's city of residence (2-128 characters)
	// @example "Moscow"
	// @MinLength(2) @MaxLength(128)
	City *string `json:"city,omitempty" validate:"omitempty,min=2,max=128"`
	// University faculty or department (optional)
	// @example "Faculty of Computer Science"
	// @MaxLength(255)
	Faculty *string `json:"faculty,omitempty" validate:"omitempty,max=255"`
	// Course or year of study (optional)
	// @example "4"
	// @MaxLength(10)
	Course *string `json:"course,omitempty" validate:"omitempty,max=10"`
	// Reference to uploaded avatar image ID (optional)
	// @example "img_abc123"
	// @MaxLength(128)
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
// UserResponse
// @Description Public user profile information
type UserResponse struct {
	// Unique user identifier
	// @example 12345
	ID types.IdType `json:"id"`
	// User's display name
	// @example "Ivan Ivanov"
	Name string `json:"name"`
	// User's email address
	// @example "user@example.com"
	Email string `json:"email"`
	// User role: "user" or "moderator"
	// @example "user"
	Role string `json:"role"`
	// Short bio or about me section
	// @example "Software engineer passionate about Go"
	About *string `json:"about,omitempty"`
	// User's city of residence
	// @example "Moscow"
	City string `json:"city"`
	// University faculty or department
	// @example "Faculty of Computer Science"
	Faculty *string `json:"faculty,omitempty"`
	// Course or year of study
	// @example "4"
	Course *string `json:"course,omitempty"`
	// Reference to avatar image ID if set
	// @example "img_abc123"
	AvatarImageID *string `json:"avatar_image_id,omitempty"`
	// Account creation timestamp in RFC3339 format
	// @example "2024-01-15T08:30:00Z"
	CreatedAt time.Time `json:"created_at"`
	// Last profile update timestamp in RFC3339 format
	// @example "2024-02-20T14:22:00Z"
	UpdatedAt time.Time `json:"updated_at"`
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
