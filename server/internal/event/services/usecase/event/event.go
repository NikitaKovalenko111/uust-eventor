package event_service

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"eventor/internal/event/contracts/user_provider"
	domain_errors "eventor/internal/event/domain/errors"
	"eventor/internal/event/domain/models"
	event_repo "eventor/internal/event/storage/repositories/event"
)

// EventService бизнес-логика работы с событиями
type EventService struct {
	eventRepo    *event_repo.EventRepo
	userProvider user_provider.UserProvider
}

// New создаёт новый сервис
func Init(eventRepo *event_repo.EventRepo, userProvider user_provider.UserProvider) *EventService {
	return &EventService{
		eventRepo:    eventRepo,
		userProvider: userProvider,
	}
}

// =====================================================
// REQUEST DTOs (для отделения транспорта от сервиса)
// =====================================================

// CreateEventRequest данные для создания события
type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	ImageID     string    `json:"image_id,omitempty"`
}

// UpdateEventRequest данные для обновления события (частичное обновление)
type UpdateEventRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	EventDate   *time.Time `json:"event_date,omitempty"`
	Location    *string    `json:"location,omitempty"`
	ImageID     *string    `json:"image_id,omitempty"`
}

// =====================================================
// CREATE
// =====================================================

// Create создаёт новое событие
func (s *EventService) Create(ctx context.Context, creatorID uint64, req *CreateEventRequest) (*models.Event, error) {
	// 1. Нормализация
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Location = strings.TrimSpace(req.Location)
	req.ImageID = strings.TrimSpace(req.ImageID)

	// 2. Валидация
	if err := validateCreateEvent(req); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrValidation, err)
	}

	// 3. Проверка существования создателя
	if err := s.verifyCreatorExists(ctx, creatorID); err != nil {
		return nil, err
	}

	// 4. Сборка доменной модели
	event := &models.Event{
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate.Truncate(24 * time.Hour), // сохраняем только дату
		Location:    req.Location,
		ImageID:     toNullString(req.ImageID),
		CreatorID:   creatorID,
	}

	// 5. Сохранение в БД
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	return event, nil
}

// =====================================================
// READ
// =====================================================

// GetByID получает событие по ID
func (s *EventService) GetByID(ctx context.Context, id uint64) (*models.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, domain_errors.ErrNotFound) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return event, nil
}

// List получает список всех событий с пагинацией
func (s *EventService) List(ctx context.Context, limit, offset int) ([]*models.Event, error) {
	events, err := s.eventRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return events, nil
}

// GetByCreatorID получает события конкретного пользователя
func (s *EventService) GetByCreatorID(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Event, error) {
	// Опционально: проверить, что пользователь существует
	events, err := s.eventRepo.GetByCreatorID(ctx, creatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return events, nil
}

// =====================================================
// UPDATE
// =====================================================

// Update обновляет событие
func (s *EventService) Update(ctx context.Context, id uint64, creatorID uint64, req *UpdateEventRequest) (*models.Event, error) {
	// 1. Получаем существующее событие
	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, domain_errors.ErrNotFound) {
			return nil, fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	// 2. Проверка прав: обновлять может только создатель (или модератор/админ)
	if existing.CreatorID != creatorID {
		return nil, fmt.Errorf("%w: you can only edit your own events", domain_errors.ErrForbidden)
	}

	// 3. Валидация частичных обновлений
	if err := validateUpdateEvent(req); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrValidation, err)
	}

	// 4. Применяем изменения
	updated := *existing // копируем структуру
	applyUpdates(&updated, req)

	// 5. Сохраняем
	if err := s.eventRepo.Update(ctx, &updated); err != nil {
		return nil, fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	return &updated, nil
}

// =====================================================
// DELETE
// =====================================================

// Delete удаляет событие
func (s *EventService) Delete(ctx context.Context, id uint64, creatorID uint64) error {
	existing, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, domain_errors.ErrNotFound) {
			return fmt.Errorf("%w: event %d", domain_errors.ErrEventNotFound, id)
		}
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}

	// Проверка прав
	if existing.CreatorID != creatorID {
		return fmt.Errorf("%w: you can only delete your own events", domain_errors.ErrForbidden)
	}

	if err := s.eventRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %w", domain_errors.ErrDatabase, err)
	}
	return nil
}

// =====================================================
// VALIDATION FUNCTIONS
// =====================================================

func validateCreateEvent(req *CreateEventRequest) error {
	// Title
	if req.Title == "" {
		return domain_errors.ErrTitleRequired
	}
	if len(req.Title) > 255 {
		return domain_errors.ErrTitleTooLong
	}
	// Description
	if req.Description == "" {
		return domain_errors.ErrDescriptionRequired
	}
	if len(req.Description) > 5000 {
		return domain_errors.ErrDescriptionTooLong
	}
	// EventDate
	if req.EventDate.IsZero() {
		return domain_errors.ErrDateRequired
	}
	// Дата не может быть в прошлом (разрешаем сегодня)
	if req.EventDate.Truncate(24 * time.Hour).Before(time.Now().Truncate(24 * time.Hour)) {
		return domain_errors.ErrDateInPast
	}
	// Location
	if req.Location == "" {
		return domain_errors.ErrLocationRequired
	}
	if len(req.Location) > 255 {
		return domain_errors.ErrLocationTooLong
	}
	// ImageID (опционально)
	if req.ImageID != "" && len(req.ImageID) > 128 {
		return domain_errors.ErrImageIDInvalid
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
		if req.EventDate.IsZero() {
			return domain_errors.ErrDateRequired
		}
		if req.EventDate.Truncate(24 * time.Hour).Before(time.Now().Truncate(24 * time.Hour)) {
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
	if req.ImageID != nil {
		*req.ImageID = strings.TrimSpace(*req.ImageID)
		if *req.ImageID != "" && len(*req.ImageID) > 128 {
			return domain_errors.ErrImageIDInvalid
		}
	}
	return nil
}

// =====================================================
// UTILS
// =====================================================

func (s *EventService) verifyCreatorExists(ctx context.Context, creatorID uint64) error {
	_, err := s.userProvider.GetByID(ctx, creatorID)
	if err != nil {
		// Если пользователя нет в БД
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
	if req.EventDate != nil {
		event.EventDate = req.EventDate.Truncate(24 * time.Hour)
	}
	if req.Location != nil {
		event.Location = *req.Location
	}
	if req.ImageID != nil {
		event.ImageID = toNullString(*req.ImageID)
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
