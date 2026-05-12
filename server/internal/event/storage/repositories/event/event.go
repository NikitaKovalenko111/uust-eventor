package event_repository

import (
	"context"
	"database/sql"
	"errors"
	domain_errors "eventor/internal/event/domain/errors"
	"eventor/internal/event/domain/models"
	"fmt"

	"github.com/lib/pq"
)

// EventRepo репозиторий для работы с таблицей events
type EventRepo struct {
	db *sql.DB
}

// New инициализирует репозиторий
func Init(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

// =====================================================
// CREATE
// =====================================================

// Create создаёт новое событие
func (r *EventRepo) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (
			title, description, event_date, location, image_id, creator_id
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		event.Title,
		event.Description,
		event.EventDate,
		event.Location,
		event.ImageID,
		event.CreatorID,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503": // foreign_key_violation
				return fmt.Errorf("%w: creator_id %d not found", domain_errors.ErrForeignKeyViolation, event.CreatorID)
			case "23502": // not_null_violation
				return fmt.Errorf("%w: missing required field", domain_errors.ErrValidation)
			}
		}
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return nil
}

// =====================================================
// READ
// =====================================================

// GetByID получает событие по ID
func (r *EventRepo) GetByID(ctx context.Context, id uint64) (*models.Event, error) {
	query := `
		SELECT id, title, description, event_date, location, image_id, 
		       creator_id, created_at, updated_at
		FROM events WHERE id = $1
	`
	var event models.Event
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.EventDate,
		&event.Location, &event.ImageID, &event.CreatorID,
		&event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return &event, nil
}

// GetByCreatorID получает список событий создателя с пагинацией
func (r *EventRepo) GetByCreatorID(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Event, error) {
	return r.queryEvents(ctx, "WHERE creator_id = $1 ORDER BY event_date DESC, created_at DESC LIMIT $2 OFFSET $3", creatorID, limit, offset)
}

// List получает все события с пагинацией (сортировка: сначала будущие, потом по дате создания)
func (r *EventRepo) List(ctx context.Context, limit, offset int) ([]*models.Event, error) {
	return r.queryEvents(ctx, "ORDER BY event_date DESC, created_at DESC LIMIT $1 OFFSET $2", limit, offset)
}

// Count возвращает общее количество событий (для пагинации)
func (r *EventRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return count, nil
}

// =====================================================
// UPDATE
// =====================================================

// Update обновляет событие (обновляет только непустые строковые поля)
func (r *EventRepo) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events SET 
			title = COALESCE(NULLIF($1, ''), title),
			description = COALESCE(NULLIF($2, ''), description),
			event_date = COALESCE(NULLIF($3, event_date), event_date),
			location = COALESCE(NULLIF($4, ''), location),
			image_id = $5,
			creator_id = COALESCE(NULLIF($6, 0), creator_id),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		event.Title,
		event.Description,
		event.EventDate,
		event.Location,
		event.ImageID,
		event.CreatorID,
		event.ID,
	).Scan(&event.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: event %d", domain_errors.ErrNotFound, event.ID)
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return fmt.Errorf("%w: invalid creator_id", domain_errors.ErrForeignKeyViolation)
		}
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return nil
}

// =====================================================
// DELETE
// =====================================================

// Delete удаляет событие по ID
func (r *EventRepo) Delete(ctx context.Context, id uint64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: event %d", domain_errors.ErrNotFound, id)
	}
	return nil
}

// =====================================================
// INTERNAL HELPERS
// =====================================================

// queryEvents универсальный метод для SELECT + пагинации
func (r *EventRepo) queryEvents(ctx context.Context, condition string, args ...interface{}) ([]*models.Event, error) {
	query := fmt.Sprintf(`
		SELECT id, title, description, event_date, location, image_id, 
		       creator_id, created_at, updated_at
		FROM events %s
	`, condition)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EventDate,
			&e.Location, &e.ImageID, &e.CreatorID,
			&e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: scan error", domain_errors.ErrDatabase)
		}
		events = append(events, &e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return events, nil
}
