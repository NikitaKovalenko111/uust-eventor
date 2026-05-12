package models

import (
	"database/sql"
	"time"
)

// Event доменная сущность события
type Event struct {
	ID          uint64         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	EventDate   time.Time      `json:"event_date"`
	Location    string         `json:"location"`
	ImageID     sql.NullString `json:"image_id,omitempty"`
	CreatorID   uint64         `json:"creator_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
