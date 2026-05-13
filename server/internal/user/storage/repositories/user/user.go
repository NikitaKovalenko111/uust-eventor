package user_storage

import (
	"context"
	"database/sql"
	"errors"
	"eventor/internal/platform/types"
	"eventor/internal/user/domain/models"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type UserRepo struct {
	db *sql.DB
}

func Init(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(ctx context.Context, u *models.User, tx *sql.Tx) error {
	query := `
		INSERT INTO users (
			name, email, password_hash, role, about, 
			city, faculty, course, avatar_image_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query,
		u.Name,
		u.Email,
		u.PasswordHash,
		u.Role,
		u.About,
		u.City,
		u.Faculty,
		u.Course,
		u.AvatarImageID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return fmt.Errorf("user with email %s already exists: %w", u.Email, err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id types.IdType) (*models.User, error) {
	query := `
		SELECT id, name, email, role, about, city, faculty, course, 
		       avatar_image_id, created_at, updated_at 
		FROM users WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.About,
		&user.City, &user.Faculty, &user.Course, &user.AvatarImageID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, about, city, faculty, 
		       course, avatar_image_id, created_at, updated_at 
		FROM users WHERE email = $1
	`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.About,
		&user.City, &user.Faculty, &user.Course, &user.AvatarImageID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET 
			name = COALESCE(NULLIF($1, ''), name),
			email = COALESCE(NULLIF($2, ''), email),
			password_hash = COALESCE(NULLIF($3, ''), password_hash),
			role = COALESCE(NULLIF($4, ''), role),
			about = $5,
			city = COALESCE(NULLIF($6, ''), city),
			faculty = $7,
			course = $8,
			avatar_image_id = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.About,
		user.City,
		user.Faculty,
		user.Course,
		user.AvatarImageID,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user with id %d not found", user.ID)
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (r *UserRepo) UpdatePartial(ctx context.Context, id types.IdType, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Разрешённые поля для обновления
	allowedFields := map[string]bool{
		"name": true, "email": true, "password_hash": true,
		"role": true, "about": true, "city": true,
		"faculty": true, "course": true, "avatar_image_id": true,
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	idx := 1

	for field, value := range updates {
		if !allowedFields[field] {
			continue // Игнорировать неизвестные поля
		}
		// Для about, faculty, course, avatar_image_id допускаем NULL
		if value == nil {
			setParts = append(setParts, fmt.Sprintf("%s = NULL", field))
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, idx))
			args = append(args, value)
			idx++
		}
	}

	if len(setParts) == 0 {
		return errors.New("no valid fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d RETURNING updated_at",
		strings.Join(setParts, ", "), idx,
	)

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user with id %d not found", id)
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id types.IdType) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}

// List получает список пользователей с пагинацией
func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, name, email, role, about, city, faculty, course, 
		       avatar_image_id, created_at, updated_at 
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Role, &user.About,
			&user.City, &user.Faculty, &user.Course, &user.AvatarImageID,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return users, nil
}
