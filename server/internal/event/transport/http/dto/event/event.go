// internal/event/transport/http/dto/event/event.go
package event

import (
	"database/sql"
	"time"

	"eventor/internal/event/domain/models"
)

type EventResponse struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	ImageID     *string   `json:"image_id,omitempty"`
	CreatorID   uint64    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToResponse(e *models.Event) *EventResponse {
	return &EventResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		EventDate:   e.EventDate,
		Location:    e.Location,
		ImageID:     nullStringToPtr(e.ImageID),
		CreatorID:   e.CreatorID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
