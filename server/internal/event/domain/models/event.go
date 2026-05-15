package models

import (
	"database/sql"
	"eventor/internal/platform/types"
	"time"
)

// Event доменная сущность события
type Event struct {
	ID          types.IdType   `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	EventDate   time.Time      `json:"event_date"`
	Location    string         `json:"location"`
	ImageID     sql.NullString `json:"image_id,omitempty"`
	CreatorID   types.IdType   `json:"creator_id"`
	Tags        []string       `json:"tags"`
	Attendees   []types.IdType `json:"attendees"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
