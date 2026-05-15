package event_service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	"eventor/internal/event/contracts/user_provider"
	domain_errors "eventor/internal/event/domain/errors"
	"eventor/internal/event/domain/models"
	event_repo "eventor/internal/event/storage/repositories/event"
	file_storage "eventor/internal/platform/storage/files"
	"eventor/internal/platform/types"
)

// EventService бизнес-логика работы с событиями
type EventService struct {
	eventRepo    *event_repo.EventRepo
	userProvider user_provider.UserProvider
	fileStorage  *file_storage.FileStorage
}

// New создаёт новый сервис
func Init(eventRepo *event_repo.EventRepo, userProvider user_provider.UserProvider, fileStorage *file_storage.FileStorage) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		userProvider: userProvider,
		fileStorage:  fileStorage,
	}
}

// =====================================================
// REQUEST DTOs (для отделения транспорта от сервиса)
// =====================================================

// CreateEventRequest данные для создания события
type CreateEventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EventDate   string   `json:"event_date"`
	Location    string   `json:"location"`
	ImageURI    string   `json:"image_uri,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateEventRequest данные для обновления события (частичное обновление)
type UpdateEventRequest struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	EventDate   *string   `json:"event_date,omitempty"`
	Location    *string   `json:"location,omitempty"`
	ImageURI    *string   `json:"image_uri,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}

// =====================================================
// CREATE
// =====================================================

// Create создаёт новое событие
func (s *EventService) Create(ctx context.Context, creatorID types.IdType, req *CreateEventRequest) (*models.Event, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)
	req.ImageURI = strings.TrimSpace(req.ImageURI)
	req.Tags = normalizeTags(req.Tags)

	if err := validateCreateEvent(req); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrValidation, err)
	}

	if err := s.verifyCreatorExists(ctx, creatorID); err != nil {
		return nil, err
	}

	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid event_date format", domain_errors.ErrValidation)
	}

	event := &models.Event{
		Title:       req.Title,
		Description: req.Description,
		EventDate:   eventDate,
		Location:    req.Location,
		ImageID:     toNullString(req.ImageURI),
		CreatorID:   creatorID,
		Tags:        req.Tags,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// =====================================================
// READ
// =====================================================

// GetByID получает событие по ID
func (s *EventService) GetByID(ctx context.Context, id types.IdType) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), domain_errors.ErrNotFound.Error()) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return nil, err
	}
	return event, nil
}

// List получает список всех событий с пагинацией
func (s *EventService) List(ctx context.Context, limit, offset int, search string) ([]*models.Event, error) {
	events, err := s.eventRepo.List(ctx, limit, offset, search)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetByCreatorID получает события конкретного пользователя
func (s *EventService) GetByCreatorID(ctx context.Context, creatorID types.IdType, limit, offset int) ([]*models.Event, error) {
	events, err := s.eventRepo.GetByCreatorID(ctx, creatorID, limit, offset)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// =====================================================
// UPDATE
// =====================================================

// Update обновляет событие
func (s *EventService) Update(ctx context.Context, id types.IdType, creatorID types.IdType, req *UpdateEventRequest) (*models.Event, error) {
	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), domain_errors.ErrNotFound.Error()) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return nil, err
	}

	if existing.CreatorID != creatorID {
		return nil, fmt.Errorf("%w: you can only edit your own events", domain_errors.ErrForbidden)
	}

	trimUpdateRequest(req)
	parsedDate, err := parseOptionalDate(req.EventDate)
	if err != nil {
		return nil, err
	}

	if err := validateUpdateEvent(req); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrValidation, err)
	}

	updated := *existing
	applyUpdates(&updated, req)
	if parsedDate != nil {
		updated.EventDate = *parsedDate
	}

	if err := s.eventRepo.Update(ctx, &updated); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, id)
}

// =====================================================
// DELETE
// =====================================================

// Delete удаляет событие
func (s *EventService) Delete(ctx context.Context, id types.IdType, creatorID types.IdType) error {
	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), domain_errors.ErrNotFound.Error()) {
			return fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return err
	}

	// Проверка прав
	if existing.CreatorID != creatorID {
		return fmt.Errorf("%w: you can only delete your own events", domain_errors.ErrForbidden)
	}

	if s.fileStorage != nil && existing.ImageID.Valid && strings.TrimSpace(existing.ImageID.String) != "" {
		if err := s.fileStorage.DeleteEventImage(ctx, existing.ImageID.String); err != nil {
			return fmt.Errorf("file storage error: %w", err)
		}
	}

	if err := s.eventRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *EventService) UploadImage(ctx context.Context, content io.Reader, size int64, contentType string) (string, error) {
	if s.fileStorage == nil {
		return "", fmt.Errorf("file storage is not initialized")
	}

	imageID, err := s.fileStorage.UploadEventImage(ctx, content, size, contentType)
	if err != nil {
		return "", fmt.Errorf("file storage error: %w", err)
	}

	return imageID, nil
}

func (s *EventService) Register(ctx context.Context, eventID, userID types.IdType) (*models.Event, error) {
	if err := s.eventRepo.RegisterAttendee(ctx, eventID, userID); err != nil {
		return nil, err
	}
	return s.eventRepo.GetByID(ctx, eventID)
}

func (s *EventService) Unregister(ctx context.Context, eventID, userID types.IdType) (*models.Event, error) {
	if err := s.eventRepo.UnregisterAttendee(ctx, eventID, userID); err != nil {
		return nil, err
	}
	return s.eventRepo.GetByID(ctx, eventID)
}

// Finish помечает событие как завершённое (только создатель может выполнить)
func (s *EventService) Finish(ctx context.Context, id types.IdType, creatorID types.IdType) (*models.Event, error) {
	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), domain_errors.ErrNotFound.Error()) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return nil, err
	}

	if existing.CreatorID != creatorID {
		return nil, fmt.Errorf("%w: you can only finish your own events", domain_errors.ErrForbidden)
	}

	if err := s.eventRepo.SetFinished(ctx, id, true); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, id)
}

// =====================================================
// VALIDATION FUNCTIONS
// =====================================================

func validateCreateEvent(req *CreateEventRequest) error {
	if req.Title == "" {
		return domain_errors.ErrTitleRequired
	}
	if len(req.Title) > 255 {
		return domain_errors.ErrTitleTooLong
	}
	if req.Description == "" {
		return domain_errors.ErrDescriptionRequired
	}
	if len(req.Description) > 5000 {
		return domain_errors.ErrDescriptionTooLong
	}
	if strings.TrimSpace(req.EventDate) == "" {
		return domain_errors.ErrDateRequired
	}
	parsedDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		return domain_errors.ErrDateRequired
	}
	if parsedDate.Truncate(24 * time.Hour).Before(time.Now().Truncate(24 * time.Hour)) {
		return domain_errors.ErrDateInPast
	}
	if req.Location == "" {
		return domain_errors.ErrLocationRequired
	}
	if len(req.Location) > 255 {
		return domain_errors.ErrLocationTooLong
	}
	if req.ImageURI != "" && len(req.ImageURI) > 512 {
		return domain_errors.ErrImageIDInvalid
	}
	for _, tag := range req.Tags {
		if len(strings.TrimSpace(tag)) > 64 {
			return domain_errors.ErrTagTooLong
		}
	}
	return nil
}

func validateUpdateEvent(req *UpdateEventRequest) error {
	if req.Title != nil {
		*req.Title = strings.TrimSpace(*req.Title)
		if *req.Title == "" {
			return domain_errors.ErrTitleRequired
		}
		if len(*req.Title) > 255 {
			return domain_errors.ErrTitleTooLong
		}
	}
	if req.Description != nil {
		*req.Description = strings.TrimSpace(*req.Description)
		if *req.Description == "" {
			return domain_errors.ErrDescriptionRequired
		}
		if len(*req.Description) > 5000 {
			return domain_errors.ErrDescriptionTooLong
		}
	}
	if req.EventDate != nil {
		parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(*req.EventDate))
		if err != nil {
			return domain_errors.ErrDateRequired
		}
		if parsedDate.Truncate(24 * time.Hour).Before(time.Now().Truncate(24 * time.Hour)) {
			return domain_errors.ErrDateInPast
		}
	}
	if req.Location != nil {
		*req.Location = strings.TrimSpace(*req.Location)
		if *req.Location == "" {
			return domain_errors.ErrLocationRequired
		}
		if len(*req.Location) > 255 {
			return domain_errors.ErrLocationTooLong
		}
	}
	if req.ImageURI != nil {
		*req.ImageURI = strings.TrimSpace(*req.ImageURI)
		if *req.ImageURI != "" && len(*req.ImageURI) > 512 {
			return domain_errors.ErrImageIDInvalid
		}
	}
	if req.Tags != nil {
		cleaned := normalizeTags(*req.Tags)
		for _, tag := range cleaned {
			if len(tag) > 64 {
				return domain_errors.ErrTagTooLong
			}
		}
		*req.Tags = cleaned
	}
	return nil
}

// =====================================================
// UTILS
// =====================================================

func (s *EventService) verifyCreatorExists(ctx context.Context, creatorID types.IdType) error {
	_, err := s.userProvider.GetByID(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("%w: creator %d does not exist", domain_errors.ErrCreatorNotFound, creatorID)
	}
	return nil
}

func applyUpdates(event *models.Event, req *UpdateEventRequest) {
	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Location != nil {
		event.Location = *req.Location
	}
	if req.ImageURI != nil {
		event.ImageID = toNullString(*req.ImageURI)
	}
	if req.Tags != nil {
		event.Tags = append([]string(nil), (*req.Tags)...)
	}
}

func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	if strings.TrimSpace(*value) == "" {
		return nil, domain_errors.ErrDateRequired
	}
	parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(*value))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid event_date format", domain_errors.ErrValidation)
	}
	return &parsedDate, nil
}

func trimUpdateRequest(req *UpdateEventRequest) {
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		req.Title = &trimmed
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		req.Description = &trimmed
	}
	if req.EventDate != nil {
		trimmed := strings.TrimSpace(*req.EventDate)
		req.EventDate = &trimmed
	}
	if req.Location != nil {
		trimmed := strings.TrimSpace(*req.Location)
		req.Location = &trimmed
	}
	if req.ImageURI != nil {
		trimmed := strings.TrimSpace(*req.ImageURI)
		req.ImageURI = &trimmed
	}
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
