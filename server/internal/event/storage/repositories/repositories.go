package repositories

import (
	"database/sql"
	event_storage "eventor/internal/event/storage/repositories/event"
)

type Repos struct {
	EventRepo *event_storage.EventRepo
}

func Init(db *sql.DB) *Repos {
	return &Repos{
		EventRepo: event_storage.Init(db),
	}
}
