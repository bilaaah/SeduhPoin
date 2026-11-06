package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

type SessionRepo struct {
	DB *sql.DB
}

func (r *SessionRepo) Buat(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES (?, ?, NOW() + INTERVAL 7 DAY)`, token, userID)

	return token, err
}

func (r *SessionRepo) UserDariToken(ctx context.Context, token string) (model.User, error) {
	var u model.User

	err := r.DB.QueryRowContext(ctx, `
		SELECT u.id, u.nama, u.email, u.role
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token = ? AND s.expires_at > NOW()`, token,
	).Scan(&u.ID, &u.Nama, &u.Email, &u.Role)

	if errors.Is(err, sql.ErrNoRows) {
		return u, model.ErrNotFound
	}

	return u, err
}

func (r *SessionRepo) Hapus(ctx context.Context, token string) error {
	_, err := r.DB.ExecContext(ctx,
		"DELETE FROM sessions WHERE token = ?", token)

	return err
}

