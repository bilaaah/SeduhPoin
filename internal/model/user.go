package model

import "errors"

var (
	ErrEmailTerpakai = errors.New("email sudah terdaftar")
	ErrLoginGagal    = errors.New("email atau password salah")
)

// User tidak memuat password hash,
// jadi aman dikirim ke handler dan template.
type User struct {
	ID    int64
	Nama  string
	Email string
	Role  string
}
