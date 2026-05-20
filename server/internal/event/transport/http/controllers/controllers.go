package controllers

import (
	"eventor/internal/event/services"
	event_controller "eventor/internal/event/transport/http/controllers/event"
	"log/slog"

	file_storage "eventor/internal/platform/storage/files"
)

type Controllers struct {
	logger          *slog.Logger
	EventController *event_controller.EventController
}

func Init(services *services.Services, fileStorage *file_storage.FileStorage, logger *slog.Logger) *Controllers {
	return &Controllers{
		logger:          logger,
		EventController: event_controller.Init(logger, services.EventService, fileStorage),
	}
}
