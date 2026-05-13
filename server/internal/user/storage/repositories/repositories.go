package repositories

import (
	"database/sql"
	user_storage "eventor/internal/user/storage/repositories/user"
)

type Repos struct {
	UserRepo *user_storage.UserRepo
	// Here are repos
}

func Init(db *sql.DB) *Repos {
	return &Repos{
		UserRepo: user_storage.Init(db),
	}
}
