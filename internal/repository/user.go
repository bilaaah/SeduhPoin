package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/bilaaah/SeduhPoin/internal/model"
)

type UserRepo struct {
	DB *sql.DB
}

func (r *UserRepo) Tambah(ctx context.Context, nama, email, hash, role string) error {
	if role == "" {
		role = "ADMIN"
	}
	_, err := r.DB.ExecContext(ctx,
		"INSERT INTO users (nama, email, password_hash, role) VALUES (?, ?, ?, ?)",
		nama, email, hash, role)

	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return model.ErrEmailTerpakai
	}

	if err != nil {
		return fmt.Errorf("tambah user: %w", err)
	}

	return nil
}

// HashByEmail hanya dipakai saat login.
func (r *UserRepo) HashByEmail(ctx context.Context, email string) (int64, string, error) {
	var id int64
	var hash string

	err := r.DB.QueryRowContext(ctx,
		"SELECT id, password_hash FROM users WHERE email = ?", email,
	).Scan(&id, &hash)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", model.ErrNotFound
	}

	return id, hash, err
}

