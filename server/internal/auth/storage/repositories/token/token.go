package token_repo

import (
	"database/sql"
	"eventor/internal/auth/domain/models"
	"eventor/internal/platform/types"
)

type TokenRepo struct {
	Db *sql.DB
}

func Init(db *sql.DB) *TokenRepo {
	return &TokenRepo{
		Db: db,
	}
}

func (r *TokenRepo) Create(token *models.Token, tx *sql.Tx) (*models.Token, error) {
	query := `
		INSERT INTO auth_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := tx.QueryRow(query, token.UserId, token.Hash, token.ExpiresAt, token.CreatedAt).Scan(&token.Id)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (r *TokenRepo) FindByUserID(userID types.IdType) ([]models.Token, error) {
	query := `
		SELECT id, user_id, token_hash, created_at, expires_at
		FROM auth_tokens
		WHERE user_id = $1
	`

	rows, err := r.Db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.Token
	for rows.Next() {
		var token models.Token
		err := rows.Scan(&token.Id, &token.UserId, &token.Hash, &token.CreatedAt, &token.ExpiresAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *TokenRepo) Delete(tokenID types.IdType) error {
	query := `DELETE FROM auth_tokens WHERE id = $1`

	_, err := r.Db.Exec(query, tokenID)

	return err
}
