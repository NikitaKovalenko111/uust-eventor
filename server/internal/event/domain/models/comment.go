package models

import (
	"time"

	"eventor/internal/platform/types"
)

type EventComment struct {
	ID         types.IdType `json:"id"`
	EventID    types.IdType `json:"event_id"`
	AuthorID   types.IdType `json:"author_id"`
	AuthorName string       `json:"author_name"`
	Text       string       `json:"text"`
	CreatedAt  time.Time    `json:"created_at"`
}
