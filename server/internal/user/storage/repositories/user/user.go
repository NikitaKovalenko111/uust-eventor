package user_storage

import (
	"context"
	"database/sql"
	"errors"
	"eventor/internal/platform/types"
	erors "eventor/internal/user/domain/errors"
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
			return nil, fmt.Errorf("%w: user with id %d", erors.ErrNotFound, id)
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
			return nil, fmt.Errorf("%w: user with email %s", erors.ErrNotFound, email)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) SearchByEmail(ctx context.Context, query string, limit int) ([]*models.User, error) {
	search := strings.TrimSpace(query)
	if search == "" {
		return []*models.User{}, nil
	}

	if limit <= 0 {
		limit = 20
	}

	querySQL := `
		SELECT id, name, email, role, about, city, faculty, course,
		       avatar_image_id, created_at, updated_at
		FROM users
		WHERE LOWER(email) LIKE '%' || LOWER($1) || '%'
		ORDER BY
			CASE
				WHEN LOWER(email) = LOWER($1) THEN 0
				WHEN LOWER(email) LIKE LOWER($1) || '%' THEN 1
				ELSE 2
			END,
			email ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, querySQL, search, limit)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Role, &user.About,
			&user.City, &user.Faculty, &user.Course, &user.AvatarImageID,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("database error: %w", err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return users, nil
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
			return fmt.Errorf("%w: user with id %d", erors.ErrNotFound, user.ID)
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
			return fmt.Errorf("%w: user with id %d", erors.ErrNotFound, id)
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
		return fmt.Errorf("%w: user with id %d", erors.ErrNotFound, id)
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

// AddFriend создает взаимную связь дружбы между двумя пользователями (вставляет две строки)
func (r *UserRepo) AddFriend(ctx context.Context, userID, friendID types.IdType) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "INSERT INTO friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, friendID); err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", friendID, userID); err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	return nil
}

// CreateFriendRequest inserts a friend request in pending state
func (r *UserRepo) CreateFriendRequest(ctx context.Context, requesterID, recipientID types.IdType, message string) (int64, error) {
	var id int64
	query := `
		INSERT INTO friend_requests (requester_id, recipient_id, message, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (requester_id, recipient_id) DO UPDATE SET message = EXCLUDED.message, status = 'pending', updated_at = NOW()
		RETURNING id
	`
	if err := r.db.QueryRowContext(ctx, query, requesterID, recipientID, message).Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	return id, nil
}

// GetIncomingFriendRequests returns pending requests for recipient
func (r *UserRepo) GetIncomingFriendRequests(ctx context.Context, recipientID types.IdType) ([]*models.User, []int64, error) {
	query := `
		SELECT fr.id, u.id, u.name, u.avatar_image_id, fr.created_at
		FROM friend_requests fr
		JOIN users u ON u.id = fr.requester_id
		WHERE fr.recipient_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, recipientID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var reqID int64
		var u models.User
		var avatar sql.NullString
		var created time.Time
		if err := rows.Scan(&reqID, &u.ID, &u.Name, &avatar, &created); err != nil {
			return nil, nil, fmt.Errorf("%w: %w", erors.ErrDatabase, err)
		}
		if avatar.Valid {
			u.AvatarImageID = avatar
		}
		users = append(users, &u)
		ids = append(ids, reqID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	return users, ids, nil
}

// UpdateFriendRequestStatus sets status to accepted/rejected/cancelled
func (r *UserRepo) UpdateFriendRequestStatus(ctx context.Context, requestID int64, status string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE friend_requests SET status = $1, updated_at = NOW() WHERE id = $2", status, requestID)
	if err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: friend request %d not found", erors.ErrNotFound, requestID)
	}
	return nil
}

// GetFriendRequestByID returns requester_id and recipient_id for a friend request
func (r *UserRepo) GetFriendRequestByID(ctx context.Context, requestID int64) (types.IdType, types.IdType, error) {
	var requesterID int64
	var recipientID int64
	query := `SELECT requester_id, recipient_id FROM friend_requests WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, query, requestID).Scan(&requesterID, &recipientID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("%w: friend request %d", erors.ErrNotFound, requestID)
		}
		return 0, 0, fmt.Errorf("%w: %w", erors.ErrDatabase, err)
	}
	return types.IdType(requesterID), types.IdType(recipientID), nil
}
