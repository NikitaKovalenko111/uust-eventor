package repositories

import (
	"database/sql"
	token_repo "eventor/internal/auth/storage/repositories/token"
)

type Repos struct {
	TokenRepo *token_repo.TokenRepo
}

func Init(db *sql.DB) *Repos {
	return &Repos{
		TokenRepo: token_repo.Init(db),
	}
}
