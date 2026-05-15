// internal/event/transport/http/dto/event/event.go
package event

import (
	"database/sql"
	"fmt"
	"time"

	"eventor/internal/event/domain/models"
	"eventor/internal/platform/types"
)

type EventResponse struct {
	ID          types.IdType `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	EventDate   time.Time    `json:"event_date"`
	Location    string       `json:"location"`
	ImageID     *string      `json:"image_id,omitempty"`
	CreatorID   types.IdType `json:"creator_id"`
	Tags        []string     `json:"tags"`
	Attendees   []string     `json:"attendees"`
	Finished    bool         `json:"finished"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
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
		Tags:        append([]string(nil), e.Tags...),
		Attendees:   idSliceToStrings(e.Attendees),
		Finished:    e.Finished,
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

func idSliceToStrings(ids []types.IdType) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, fmt.Sprintf("%d", id))
	}
	return result
}
