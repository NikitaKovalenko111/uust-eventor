package event_repository

import (
	"context"
	"database/sql"
	"errors"
	domain_errors "eventor/internal/event/domain/errors"
	"eventor/internal/event/domain/models"
	"eventor/internal/platform/types"
	"fmt"
	"strings"
	"unicode"

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

func (r *EventRepo) Create(ctx context.Context, event *models.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO events (
			title, description, event_date, location, image_id, creator_id
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
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
			case "23503":
				return fmt.Errorf("%w: creator_id %d not found", domain_errors.ErrForeignKeyViolation, event.CreatorID)
			case "23502":
				return fmt.Errorf("%w: missing required field", domain_errors.ErrValidation)
			}
		}
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	if err := replaceTagsTx(ctx, tx, event.ID, event.Tags); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	return nil
}

func (r *EventRepo) GetByID(ctx context.Context, id types.IdType) (*models.Event, error) {
	events, err := r.queryEvents(ctx, "WHERE e.id = $1", "", id)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: event %d", domain_errors.ErrNotFound, id)
	}
	return events[0], nil
}

func (r *EventRepo) GetByCreatorID(ctx context.Context, creatorID types.IdType, limit, offset int) ([]*models.Event, error) {
	return r.queryEvents(ctx, "WHERE e.creator_id = $1", "ORDER BY e.event_date DESC, e.created_at DESC LIMIT $2 OFFSET $3", creatorID, limit, offset)
}

func (r *EventRepo) List(ctx context.Context, limit, offset int, search, city string) ([]*models.Event, error) {
	search = strings.TrimSpace(search)
	city = strings.TrimSpace(city)
	if search == "" {
		if city == "" {
			return r.queryEvents(ctx, "", "ORDER BY e.event_date DESC, e.created_at DESC LIMIT $1 OFFSET $2", limit, offset)
		}

		return r.queryEvents(
			ctx,
			"",
			"ORDER BY CASE WHEN COALESCE(e.location, '') ILIKE '%' || $1 || '%' THEN 0 ELSE 1 END, e.event_date DESC, e.created_at DESC LIMIT $2 OFFSET $3",
			city,
			limit,
			offset,
		)
	}

	// build a prefix tsquery from search terms: each token becomes token:* and joined with &
	tsQuery := buildPrefixTsQuery(search)
	// always include an ILIKE-based substring match (slower but covers arbitrary substrings)
	// use both FTS prefix matching and substring ILIKE; pass tsQuery as $1 and plain search as $2
	if tsQuery == "" {
		// fallback to websearch style when we couldn't build a prefix query
		whereClause := `
		WHERE (
			to_tsvector('simple',
				COALESCE(e.title, '') || ' ' ||
				COALESCE(e.description, '') || ' ' ||
				COALESCE(e.location, '') || ' ' ||
				COALESCE(e.event_date::text, '') || ' ' ||
				COALESCE((SELECT string_agg(et2.tag, ' ') FROM event_tags et2 WHERE et2.event_id = e.id), '')
			) @@ websearch_to_tsquery('simple', $1)
			OR (
				COALESCE(e.title, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.description, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.location, '') ILIKE '%' || $2 || '%' OR
				COALESCE((SELECT string_agg(et3.tag, ' ') FROM event_tags et3 WHERE et3.event_id = e.id), '') ILIKE '%' || $2 || '%'
			)
		)
	`
		return r.queryEvents(ctx, whereClause, "ORDER BY e.event_date DESC, e.created_at DESC LIMIT $3 OFFSET $4", search, search, limit, offset)
	}

	whereClause := `
		WHERE (
			to_tsvector('simple',
				COALESCE(e.title, '') || ' ' ||
				COALESCE(e.description, '') || ' ' ||
				COALESCE(e.location, '') || ' ' ||
				COALESCE(e.event_date::text, '') || ' ' ||
				COALESCE((SELECT string_agg(et2.tag, ' ') FROM event_tags et2 WHERE et2.event_id = e.id), '')
			) @@ to_tsquery('simple', $1)
			OR (
				COALESCE(e.title, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.description, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.location, '') ILIKE '%' || $2 || '%' OR
				COALESCE((SELECT string_agg(et3.tag, ' ') FROM event_tags et3 WHERE et3.event_id = e.id), '') ILIKE '%' || $2 || '%'
			)
		)
	`

	if city == "" {
		return r.queryEvents(ctx, whereClause, "ORDER BY e.event_date DESC, e.created_at DESC LIMIT $3 OFFSET $4", tsQuery, search, limit, offset)
	}

	return r.queryEvents(
		ctx,
		whereClause,
		"ORDER BY CASE WHEN COALESCE(e.location, '') ILIKE '%' || $1 || '%' THEN 0 ELSE 1 END, e.event_date DESC, e.created_at DESC LIMIT $4 OFFSET $5",
		city,
		tsQuery,
		search,
		limit,
		offset,
	)
}

func (r *EventRepo) ListAll(ctx context.Context, search string) ([]*models.Event, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return r.queryEvents(ctx, "", "ORDER BY e.event_date DESC, e.created_at DESC")
	}

	tsQuery := buildPrefixTsQuery(search)
	if tsQuery == "" {
		whereClause := `
		WHERE (
			to_tsvector('simple',
				COALESCE(e.title, '') || ' ' ||
				COALESCE(e.description, '') || ' ' ||
				COALESCE(e.location, '') || ' ' ||
				COALESCE(e.event_date::text, '') || ' ' ||
				COALESCE((SELECT string_agg(et2.tag, ' ') FROM event_tags et2 WHERE et2.event_id = e.id), '')
			) @@ websearch_to_tsquery('simple', $1)
			OR (
				COALESCE(e.title, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.description, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.location, '') ILIKE '%' || $2 || '%' OR
				COALESCE((SELECT string_agg(et3.tag, ' ') FROM event_tags et3 WHERE et3.event_id = e.id), '') ILIKE '%' || $2 || '%'
			)
		)
	`
		return r.queryEvents(ctx, whereClause, "ORDER BY e.event_date DESC, e.created_at DESC", search, search)
	}

	whereClause := `
		WHERE (
			to_tsvector('simple',
				COALESCE(e.title, '') || ' ' ||
				COALESCE(e.description, '') || ' ' ||
				COALESCE(e.location, '') || ' ' ||
				COALESCE(e.event_date::text, '') || ' ' ||
				COALESCE((SELECT string_agg(et2.tag, ' ') FROM event_tags et2 WHERE et2.event_id = e.id), '')
			) @@ to_tsquery('simple', $1)
			OR (
				COALESCE(e.title, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.description, '') ILIKE '%' || $2 || '%' OR
				COALESCE(e.location, '') ILIKE '%' || $2 || '%' OR
				COALESCE((SELECT string_agg(et3.tag, ' ') FROM event_tags et3 WHERE et3.event_id = e.id), '') ILIKE '%' || $2 || '%'
			)
		)
	`

	return r.queryEvents(ctx, whereClause, "ORDER BY e.event_date DESC, e.created_at DESC", tsQuery, search)
}

func (r *EventRepo) GetUserTagPreferences(ctx context.Context, userID types.IdType) (map[string]int, error) {
	query := `
		SELECT et.tag, COUNT(*)
		FROM event_attendees ea
		JOIN event_tags et ON et.event_id = ea.event_id
		WHERE ea.user_id = $1
		GROUP BY et.tag
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer rows.Close()

	preferences := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
		}
		preferences[strings.ToLower(strings.TrimSpace(tag))] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return preferences, nil
}

// buildPrefixTsQuery converts a free-form search string into a tsquery that uses prefix
// matching: each token becomes token:* and tokens are ANDed. Returns empty string if
// no valid tokens found.
func buildPrefixTsQuery(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		t := strings.TrimSpace(f)
		if t == "" {
			continue
		}
		// remove single quotes to avoid tsquery syntax issues
		t = strings.ReplaceAll(t, "'", "")
		parts = append(parts, t+":*")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " & ")
}

func (r *EventRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return count, nil
}

func (r *EventRepo) Update(ctx context.Context, event *models.Event) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer tx.Rollback()

	query := `
		UPDATE events SET
			title = $1,
			description = $2,
			event_date = $3,
			location = $4,
			image_id = $5,
			finished = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		event.Title,
		event.Description,
		event.EventDate,
		event.Location,
		event.ImageID,
		event.Finished,
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

	if err := replaceTagsTx(ctx, tx, event.ID, event.Tags); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	return nil
}

func (r *EventRepo) Delete(ctx context.Context, id types.IdType) error {
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

// SetFinished устанавливает признак finished у события
func (r *EventRepo) SetFinished(ctx context.Context, id types.IdType, finished bool) error {
	res, err := r.db.ExecContext(ctx, "UPDATE events SET finished = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", finished, id)
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

func (r *EventRepo) RegisterAttendee(ctx context.Context, eventID, userID types.IdType) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"INSERT INTO event_attendees (event_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		eventID, userID,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	if rows == 0 {
		return domain_errors.ErrEventAlreadyRegistered
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return nil
}

func (r *EventRepo) UnregisterAttendee(ctx context.Context, eventID, userID types.IdType) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "DELETE FROM event_attendees WHERE event_id = $1 AND user_id = $2", eventID, userID)
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	if rows == 0 {
		return domain_errors.ErrEventNotRegistered
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return nil
}

func (r *EventRepo) ListComments(ctx context.Context, eventID types.IdType) ([]*models.EventComment, error) {
	query := `
		SELECT ec.id, ec.event_id, ec.user_id, COALESCE(u.name, ''), ec.text, ec.created_at
		FROM event_comments ec
		JOIN users u ON u.id = ec.user_id
		WHERE ec.event_id = $1
		ORDER BY ec.created_at ASC, ec.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer rows.Close()

	comments := make([]*models.EventComment, 0)
	for rows.Next() {
		var comment models.EventComment
		if err := rows.Scan(&comment.ID, &comment.EventID, &comment.AuthorID, &comment.AuthorName, &comment.Text, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("%w: comment scan error: %v", domain_errors.ErrDatabase, err)
		}
		comments = append(comments, &comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return comments, nil
}

func (r *EventRepo) ListAttendees(ctx context.Context, eventID types.IdType) ([]*models.EventAttendee, error) {
	query := `
		SELECT ea.id, u.id, COALESCE(u.name, ''), COALESCE(u.avatar_image_id, ''), ea.registered_at
		FROM event_attendees ea
		JOIN users u ON u.id = ea.user_id
		WHERE ea.event_id = $1
		ORDER BY u.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer rows.Close()

	attendees := make([]*models.EventAttendee, 0)
	for rows.Next() {
		var a models.EventAttendee
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.AvatarID, &a.JoinedAt); err != nil {
			return nil, fmt.Errorf("%w: attendee scan error: %v", domain_errors.ErrDatabase, err)
		}
		attendees = append(attendees, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return attendees, nil
}

// CountFriendsRegistered returns how many of user's friends are registered for the given event.
func (r *EventRepo) CountFriendsRegistered(ctx context.Context, eventID, userID types.IdType) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM event_attendees ea
		JOIN friends f ON f.friend_id = ea.user_id
		WHERE f.user_id = $2 AND ea.event_id = $1
	`
	if err := r.db.QueryRowContext(ctx, query, eventID, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return count, nil
}

func (r *EventRepo) AddComment(ctx context.Context, eventID, authorID types.IdType, text string) (*models.EventComment, error) {
	query := `
		INSERT INTO event_comments (event_id, user_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	var comment models.EventComment
	comment.EventID = eventID
	comment.AuthorID = authorID
	comment.Text = text
	if err := r.db.QueryRowContext(ctx, query, eventID, authorID, text).Scan(&comment.ID, &comment.CreatedAt); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503":
				return nil, fmt.Errorf("%w: event or author not found", domain_errors.ErrForeignKeyViolation)
			case "23502":
				return nil, fmt.Errorf("%w: missing required field", domain_errors.ErrValidation)
			}
		}
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return &comment, nil
}

func (r *EventRepo) queryEvents(ctx context.Context, whereClause string, orderClause string, args ...interface{}) ([]*models.Event, error) {
	query := fmt.Sprintf(`
		SELECT
			e.id,
			e.title,
			e.description,
			e.event_date,
			e.location,
			e.image_id,
			e.creator_id,
			e.finished,
			e.created_at,
			e.updated_at,
			COALESCE(array_agg(DISTINCT et.tag) FILTER (WHERE et.tag IS NOT NULL), '{}') AS tags,
			COALESCE(array_agg(DISTINCT ea.user_id) FILTER (WHERE ea.user_id IS NOT NULL), '{}') AS attendees
		FROM events e
		LEFT JOIN event_tags et ON et.event_id = e.id
		LEFT JOIN event_attendees ea ON ea.event_id = e.id
		%s
		GROUP BY e.id
		%s
	`, whereClause, orderClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		var tags []string
		var attendeeIDs []int64
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&event.Location,
			&event.ImageID,
			&event.CreatorID,
			&event.Finished,
			&event.CreatedAt,
			&event.UpdatedAt,
			pq.Array(&tags),
			pq.Array(&attendeeIDs),
		); err != nil {
			return nil, fmt.Errorf("%w: scan error: %v", domain_errors.ErrDatabase, err)
		}
		event.Tags = tags
		event.Attendees = make([]types.IdType, 0, len(attendeeIDs))
		for _, attendeeID := range attendeeIDs {
			if attendeeID < 0 {
				continue
			}
			event.Attendees = append(event.Attendees, types.IdType(attendeeID))
		}
		events = append(events, &event)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return events, nil
}

func replaceTagsTx(ctx context.Context, tx *sql.Tx, eventID types.IdType, tags []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM event_tags WHERE event_id = $1", eventID); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	for _, tag := range tags {
		normalizedTag := strings.TrimSpace(tag)
		if normalizedTag == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO event_tags (event_id, tag) VALUES ($1, $2)", eventID, normalizedTag); err != nil {
			return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
		}
	}

	return nil
}
