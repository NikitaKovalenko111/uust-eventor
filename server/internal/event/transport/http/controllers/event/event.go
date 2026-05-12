// internal/event/transport/http/controller/event.go
package event_controller

import (
	"log/slog"
	"strconv"

	stderrors "errors"

	domain_errors "eventor/internal/event/domain/errors"
	event_service "eventor/internal/event/services/usecase/event"
	event_dto "eventor/internal/event/transport/http/dto/event"
	http_helpers "eventor/internal/platform/pkg/http/helpers" // ваш shared-хелпер валидации

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// EventController HTTP-контроллер для управления событиями
type EventController struct {
	logger       *slog.Logger
	validator    *validator.Validate
	eventService *event_service.EventService
}

// New создаёт контроллер
func Init(logger *slog.Logger, eventService *event_service.EventService) *EventController {
	return &EventController{
		logger:       logger,
		validator:    validator.New(),
		eventService: eventService,
	}
}

// RegisterRoutes регистрирует маршруты событий (все защищены authMiddleware)
func (c *EventController) RegisterRoutes(app *fiber.App, rout string, authMiddleware fiber.Handler) {
	router := app.Group(rout, authMiddleware)

	router.Post("/", c.CreateEvent)
	router.Get("/", c.ListEvents)
	router.Get("/:id", c.GetEvent)
	router.Put("/:id", c.UpdateEvent)
	router.Delete("/:id", c.DeleteEvent)
}

// =====================================================
// CREATE
// =====================================================

// CreateEvent godoc
//
//	@Summary		Create new event
//	@Description	Create an event by authenticated user
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			request	body		event_service.CreateEventRequest	true	"Event data"
//	@Success		201		{object}	event_dto.EventResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events [post]
func (c *EventController) CreateEvent(ctx *fiber.Ctx) error {
	creatorID, ok := ctx.Locals("user_id").(uint64)
	if !ok || creatorID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error: "unauthorized", Code: fiber.StatusUnauthorized,
		})
	}

	var req event_service.CreateEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body", Code: fiber.StatusBadRequest,
		})
	}

	if err := c.validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: http_helpers.FormatValidationErrors(err),
		})
	}

	event, err := c.eventService.Create(ctx.UserContext(), creatorID, &req)
	if err != nil {
		return c.handleServiceError(ctx, err, "create event")
	}

	return ctx.Status(fiber.StatusCreated).JSON(event_dto.ToResponse(event))
}

// =====================================================
// READ
// =====================================================

// GetEvent godoc
//
//	@Summary		Get event by ID
//	@Description	Returns event details
//	@Tags			events
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Success		200	{object}	event_dto.EventResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events/{id} [get]
func (c *EventController) GetEvent(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	event, err := c.eventService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return c.handleServiceError(ctx, err, "get event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

// ListEvents godoc
//
//	@Summary		List events
//	@Description	Get paginated list of events
//	@Tags			events
//	@Produce		json
//	@Param			limit	query		int	false	"Items per page"	default(20)
//	@Param			offset	query		int	false	"Offset"			default(0)
//	@Success		200		{object}	ListResponse
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events [get]
func (c *EventController) ListEvents(ctx *fiber.Ctx) error {
	limit := ctx.QueryInt("limit", 20)
	offset := ctx.QueryInt("offset", 0)

	if limit < 1 || limit > 100 {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "limit must be between 1 and 100", Code: fiber.StatusBadRequest,
		})
	}
	if offset < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "offset must be non-negative", Code: fiber.StatusBadRequest,
		})
	}

	events, err := c.eventService.List(ctx.UserContext(), limit, offset)
	if err != nil {
		return c.handleServiceError(ctx, err, "list events")
	}

	responses := make([]*event_dto.EventResponse, 0, len(events))
	for _, e := range events {
		responses = append(responses, event_dto.ToResponse(e))
	}

	return ctx.Status(fiber.StatusOK).JSON(ListResponse{
		Events: responses,
		Count:  len(responses),
		Limit:  limit,
		Offset: offset,
	})
}

// =====================================================
// UPDATE
// =====================================================

// UpdateEvent godoc
//
//	@Summary		Update event
//	@Description	Update event (only creator can modify)
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"Event ID"
//	@Param			request	body		event_service.UpdateEventRequest	true	"Update data"
//	@Success		200		{object}	event_dto.EventResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse	"Forbidden"
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events/{id} [put]
func (c *EventController) UpdateEvent(ctx *fiber.Ctx) error {
	creatorID, _ := ctx.Locals("user_id").(uint64)
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	var req event_service.UpdateEventRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body", Code: fiber.StatusBadRequest,
		})
	}

	if err := c.validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: http_helpers.FormatValidationErrors(err),
		})
	}

	event, err := c.eventService.Update(ctx.UserContext(), id, creatorID, &req)
	if err != nil {
		return c.handleServiceError(ctx, err, "update event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

// =====================================================
// DELETE
// =====================================================

// DeleteEvent godoc
//
//	@Summary		Delete event
//	@Description	Delete event (only creator can delete)
//	@Tags			events
//	@Produce		json
//	@Param			id	path		int	true	"Event ID"
//	@Success		204	"Deleted"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse	"Forbidden"
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events/{id} [delete]
func (c *EventController) DeleteEvent(ctx *fiber.Ctx) error {
	creatorID, _ := ctx.Locals("user_id").(uint64)
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	if err := c.eventService.Delete(ctx.UserContext(), id, creatorID); err != nil {
		return c.handleServiceError(ctx, err, "delete event")
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// =====================================================
// ERROR HANDLING & RESPONSES
// =====================================================

func (c *EventController) handleServiceError(ctx *fiber.Ctx, err error, operation string) error {
	c.logger.Error("event service error", "operation", operation, "error", err)

	switch {
	case stderrors.Is(err, domain_errors.ErrValidation):
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: http_helpers.ExtractValidationErrorDetails(err),
		})
	case stderrors.Is(err, domain_errors.ErrEventNotFound):
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "event not found",
			Code:  fiber.StatusNotFound,
		})
	case stderrors.Is(err, domain_errors.ErrForbidden):
		return ctx.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "you can only modify your own events",
			Code:  fiber.StatusForbidden,
		})
	case stderrors.Is(err, domain_errors.ErrForeignKeyViolation):
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "creator account no longer exists",
			Code:  fiber.StatusBadRequest,
		})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "internal server error",
			Code:  fiber.StatusInternalServerError,
		})
	}
}

// ErrorResponse стандартный формат ошибки
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// ListResponse формат ответа со списком
type ListResponse struct {
	Events []*event_dto.EventResponse `json:"events"`
	Count  int                        `json:"count"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}
