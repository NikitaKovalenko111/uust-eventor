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
	ID             types.IdType `json:"id"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	EventDate      time.Time    `json:"event_date"`
	Location       string       `json:"location"`
	ImageID        *string      `json:"image_id,omitempty"`
	CreatorID      types.IdType `json:"creator_id"`
	Tags           []string     `json:"tags"`
	Attendees      []string     `json:"attendees"`
	Finished       bool         `json:"finished"`
	RelevanceScore int          `json:"relevance_score"`
	FriendsCount   int          `json:"friends_count"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type CommentResponse struct {
	ID         types.IdType `json:"id"`
	EventID    types.IdType `json:"event_id"`
	AuthorID   types.IdType `json:"author_id"`
	AuthorName string       `json:"author_name"`
	Text       string       `json:"text"`
	CreatedAt  time.Time    `json:"created_at"`
}

type CommentListResponse struct {
	Comments []*CommentResponse `json:"comments"`
	Count    int                `json:"count"`
}

type CreateCommentRequest struct {
	Text string `json:"text" validate:"required,max=2000"`
}

func ToResponse(e *models.Event) *EventResponse {
	return &EventResponse{
		ID:             e.ID,
		Title:          e.Title,
		Description:    e.Description,
		EventDate:      e.EventDate,
		Location:       e.Location,
		ImageID:        nullStringToPtr(e.ImageID),
		CreatorID:      e.CreatorID,
		Tags:           append([]string(nil), e.Tags...),
		Attendees:      idSliceToStrings(e.Attendees),
		Finished:       e.Finished,
		RelevanceScore: 0,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func ToRankedResponse(e *models.Event, relevanceScore int) *EventResponse {
	resp := ToResponse(e)
	resp.RelevanceScore = relevanceScore
	// default friends count is zero; service may set it later if available
	return resp
}

type AttendeeResponse struct {
	ID       types.IdType `json:"id"`
	UserID   types.IdType `json:"user_id"`
	Name     string       `json:"name"`
	AvatarID string       `json:"avatar_image_id,omitempty"`
	JoinedAt time.Time    `json:"joined_at"`
}

type AttendeeListResponse struct {
	Attendees []*AttendeeResponse `json:"attendees"`
	Count     int                 `json:"count"`
}

func AttendeesToListResponse(att []*models.EventAttendee) *AttendeeListResponse {
	res := &AttendeeListResponse{
		Attendees: make([]*AttendeeResponse, 0, len(att)),
		Count:     len(att),
	}
	for _, a := range att {
		res.Attendees = append(res.Attendees, &AttendeeResponse{
			ID:       a.ID,
			UserID:   a.UserID,
			Name:     a.Name,
			AvatarID: a.AvatarID,
			JoinedAt: a.JoinedAt,
		})
	}
	return res
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
