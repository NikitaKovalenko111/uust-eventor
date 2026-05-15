// internal/event/transport/http/controller/event.go
package event_controller

import (
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	stderrors "errors"

	domain_errors "eventor/internal/event/domain/errors"
	event_service "eventor/internal/event/services/usecase/event"
	event_dto "eventor/internal/event/transport/http/dto/event"
	http_helpers "eventor/internal/platform/pkg/http/helpers" // ваш shared-хелпер валидации
	file_storage "eventor/internal/platform/storage/files"
	"eventor/internal/platform/types"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// EventController HTTP-контроллер для управления событиями
type EventController struct {
	logger       *slog.Logger
	validator    *validator.Validate
	eventService *event_service.EventService
	fileStorage  *file_storage.FileStorage
}

// New создаёт контроллер
func Init(logger *slog.Logger, eventService *event_service.EventService, fileStorage *file_storage.FileStorage) *EventController {
	return &EventController{
		logger:       logger,
		validator:    validator.New(),
		eventService: eventService,
		fileStorage:  fileStorage,
	}
}

// RegisterRoutes регистрирует маршруты событий (все защищены authMiddleware)
func (c *EventController) RegisterRoutes(app *fiber.App, rout string, authMiddleware *fiber.Handler) {
	publicRouter := app.Group(rout)

	publicRouter.Get("", c.ListEvents)
	publicRouter.Get("/", c.ListEvents)
	publicRouter.Get("/:id/comments", c.ListComments)
	publicRouter.Get("/:id", c.GetEvent)
	publicRouter.Get("/images/:image_id/file", c.GetImageFile)

	app.Post(rout, *authMiddleware, c.CreateEvent)
	app.Post(rout+"/", *authMiddleware, c.CreateEvent)
	app.Post(rout+"/image", *authMiddleware, c.UploadImage)
	app.Post(rout+"/:id/comments", *authMiddleware, c.CreateComment)
	app.Put(rout+"/:id", *authMiddleware, c.UpdateEvent)
	app.Delete(rout+"/:id", *authMiddleware, c.DeleteEvent)
	app.Post(rout+"/:id/finish", *authMiddleware, c.FinishEvent)
	app.Post(rout+"/:id/register", *authMiddleware, c.RegisterEvent)
	app.Delete(rout+"/:id/register", *authMiddleware, c.UnregisterEvent)
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
	creatorID, ok := ctx.Locals("user_id").(types.IdType)
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

// UploadImage godoc
//
//	@Summary		Upload event cover image
//	@Description	Upload image for event cover and receive storage key
//	@Tags			events
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			image	formData	file	true	"Cover image"
//	@Success		200	{object}	EventImageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/events/image [post]
func (c *EventController) UploadImage(ctx *fiber.Ctx) error {
	if c.fileStorage == nil {
		return ctx.Status(http.StatusInternalServerError).JSON(ErrorResponse{Error: "file storage is not initialized", Code: http.StatusInternalServerError})
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "image file is required", Code: http.StatusBadRequest})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(fileHeader.Filename)))
	}
	if contentType == "" {
		contentType = "image/jpeg"
	}
	if !strings.HasPrefix(contentType, "image/") {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "image must be an image", Code: http.StatusBadRequest})
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.logger.Error("failed to open event image file", slog.Any("error", err))
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "cannot read image file", Code: http.StatusBadRequest})
	}
	defer file.Close()

	imageID, err := c.eventService.UploadImage(ctx.UserContext(), file, fileHeader.Size, contentType)
	if err != nil {
		return c.handleServiceError(ctx, err, "upload event image")
	}

	return ctx.Status(http.StatusOK).JSON(EventImageResponse{
		ImageID: imageID,
		URL:     "/api/v1/events/images/" + imageID + "/file",
	})
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
	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 32)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	id := types.IdType(idParam)

	event, err := c.eventService.GetByID(ctx.UserContext(), id)
	if err != nil {
		return c.handleServiceError(ctx, err, "get event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

func (c *EventController) GetImageFile(ctx *fiber.Ctx) error {
	if c.fileStorage == nil {
		return ctx.Status(http.StatusInternalServerError).JSON(ErrorResponse{Error: "file storage is not initialized", Code: http.StatusInternalServerError})
	}

	imageID := strings.TrimSpace(ctx.Params("image_id"))
	if imageID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "invalid image ID", Code: http.StatusBadRequest})
	}

	reader, contentType, err := c.fileStorage.OpenEventImage(ctx.UserContext(), imageID)
	if err != nil {
		return c.handleServiceError(ctx, err, "get event image")
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	imageBytes, err := io.ReadAll(reader)
	if err != nil {
		c.logger.Error("failed to read event image file", slog.Any("error", err), slog.String("image_id", imageID))
		return ctx.Status(http.StatusInternalServerError).JSON(ErrorResponse{Error: "cannot read event image file", Code: http.StatusInternalServerError})
	}

	ctx.Set(fiber.HeaderContentType, contentType)
	return ctx.Send(imageBytes)
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
	search := strings.TrimSpace(ctx.Query("search"))

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

	events, err := c.eventService.List(ctx.UserContext(), limit, offset, search)
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

func (c *EventController) ListComments(ctx *fiber.Ctx) error {
	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid event ID", Code: fiber.StatusBadRequest})
	}

	comments, err := c.eventService.ListComments(ctx.UserContext(), types.IdType(idParam))
	if err != nil {
		return c.handleServiceError(ctx, err, "list comments")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.CommentsToListResponse(comments))
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
	creatorID, _ := ctx.Locals("user_id").(types.IdType)
	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	id := types.IdType(idParam)

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
	creatorID, _ := ctx.Locals("user_id").(types.IdType)
	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	id := types.IdType(idParam)

	if err := c.eventService.Delete(ctx.UserContext(), id, creatorID); err != nil {
		return c.handleServiceError(ctx, err, "delete event")
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *EventController) RegisterEvent(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok || userID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized", Code: fiber.StatusUnauthorized})
	}

	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid event ID", Code: fiber.StatusBadRequest})
	}

	event, err := c.eventService.Register(ctx.UserContext(), types.IdType(idParam), userID)
	if err != nil {
		return c.handleServiceError(ctx, err, "register event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

func (c *EventController) UnregisterEvent(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok || userID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized", Code: fiber.StatusUnauthorized})
	}

	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid event ID", Code: fiber.StatusBadRequest})
	}

	event, err := c.eventService.Unregister(ctx.UserContext(), types.IdType(idParam), userID)
	if err != nil {
		return c.handleServiceError(ctx, err, "unregister event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

// FinishEvent godoc
//
// @Summary Finish event
// @Description Mark event as finished (only creator)
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} event_dto.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/events/{id}/finish [post]
func (c *EventController) FinishEvent(ctx *fiber.Ctx) error {
	creatorID, _ := ctx.Locals("user_id").(types.IdType)
	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid event ID", Code: fiber.StatusBadRequest,
		})
	}

	event, err := c.eventService.Finish(ctx.UserContext(), types.IdType(idParam), creatorID)
	if err != nil {
		return c.handleServiceError(ctx, err, "finish event")
	}

	return ctx.Status(fiber.StatusOK).JSON(event_dto.ToResponse(event))
}

func (c *EventController) CreateComment(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok || userID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized", Code: fiber.StatusUnauthorized})
	}

	idParam, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid event ID", Code: fiber.StatusBadRequest})
	}

	var req event_dto.CreateCommentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body", Code: fiber.StatusBadRequest})
	}

	if err := c.validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: http_helpers.FormatValidationErrors(err),
		})
	}

	comment, err := c.eventService.AddComment(ctx.UserContext(), types.IdType(idParam), userID, &event_service.CreateCommentRequest{Text: req.Text})
	if err != nil {
		return c.handleServiceError(ctx, err, "create comment")
	}

	return ctx.Status(fiber.StatusCreated).JSON(event_dto.CommentToResponse(comment))
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
	case stderrors.Is(err, domain_errors.ErrEventAlreadyRegistered):
		return ctx.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Error: "event already registered",
			Code:  fiber.StatusConflict,
		})
	case stderrors.Is(err, domain_errors.ErrEventNotRegistered):
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "event registration not found",
			Code:  fiber.StatusNotFound,
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

type EventImageResponse struct {
	ImageID string `json:"image_id"`
	URL     string `json:"url"`
}
