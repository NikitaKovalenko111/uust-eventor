// internal/event/transport/http/dto/event/event.go
package event

import (
	"database/sql"
	"fmt"
	"time"

	"eventor/internal/event/domain/models"
	"eventor/internal/platform/types"
)

// EventResponse
// @Description Full event information returned by the API
type EventResponse struct {
	// Unique event identifier
	// @example 12345
	ID types.IdType `json:"id"`
	// Event title
	// @example "Tech Conference 2024"
	Title string `json:"title"`
	// Event description
	// @example "Annual conference about latest tech trends"
	Description string `json:"description"`
	// Event date and time in RFC3339 format
	// @example "2024-12-15T10:00:00Z"
	EventDate time.Time `json:"event_date"`
	// Event location
	// @example "Moscow, Tverskaya 15"
	Location string `json:"location"`
	// Image ID if cover image is set
	// @example "img_abc123xyz"
	ImageID *string `json:"image_id,omitempty"`
	// Creator user ID
	// @example 67890
	CreatorID types.IdType `json:"creator_id"`
	// Event tags
	// @example ["technology","conference"]]
	Tags []string `json:"tags"`
	// List of registered attendee user IDs (as strings)
	// @example ["67890","67891"]
	Attendees []string `json:"attendees"`
	// Whether event is marked as finished
	// @example false
	Finished bool `json:"finished"`
	// Creation timestamp in RFC3339 format
	// @example "2024-01-15T08:30:00Z"
	CreatedAt time.Time `json:"created_at"`
	// Last update timestamp in RFC3339 format
	// @example "2024-01-20T14:22:00Z"
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentResponse
// @Description Comment information
type CommentResponse struct {
	// Unique comment identifier
	// @example 999
	ID types.IdType `json:"id"`
	// Associated event ID
	// @example 12345
	EventID types.IdType `json:"event_id"`
	// Author user ID
	// @example 67890
	AuthorID types.IdType `json:"author_id"`
	// Author display name
	// @example "Ivan Ivanov"
	AuthorName string `json:"author_name"`
	// Comment text
	// @example "Great event!"
	Text string `json:"text"`
	// Creation timestamp in RFC3339 format
	// @example "2024-01-16T10:15:00Z"
	CreatedAt time.Time `json:"created_at"`
}

// CommentListResponse
// @Description Paginated list of comments
type CommentListResponse struct {
	// List of comments
	Comments []*CommentResponse `json:"comments"`
	// Total number of comments
	// @example 42
	Count int `json:"count"`
}

type CreateCommentRequest struct {
	Text string `json:"text" validate:"required,max=2000"`
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

func CommentToResponse(comment *models.EventComment) *CommentResponse {
	return &CommentResponse{
		ID:         comment.ID,
		EventID:    comment.EventID,
		AuthorID:   comment.AuthorID,
		AuthorName: comment.AuthorName,
		Text:       comment.Text,
		CreatedAt:  comment.CreatedAt,
	}
}

func CommentsToListResponse(comments []*models.EventComment) *CommentListResponse {
	result := &CommentListResponse{
		Comments: make([]*CommentResponse, 0, len(comments)),
		Count:    len(comments),
	}
	for _, comment := range comments {
		result.Comments = append(result.Comments, CommentToResponse(comment))
	}
	return result
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
