package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type ValidasiError struct {
	Pesan string
}

func (e *ValidasiError) Error() string {
	return e.Pesan
}

type AuthService struct {
	Users    *repository.UserRepo
	Sessions *repository.SessionRepo
}

func (s *AuthService) Register(ctx context.Context, nama, email, password, role string) error {
	nama = strings.TrimSpace(nama)
	email = strings.ToLower(strings.TrimSpace(email))
	role = strings.ToUpper(strings.TrimSpace(role))

	if nama == "" {
		return &ValidasiError{"Nama wajib diisi"}
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return &ValidasiError{"Format email tidak valid"}
	}

	if len(password) < 8 || len(password) > 72 {
		return &ValidasiError{"Password 8 sampai 72 karakter"}
	}

	if role != "ADMIN" && role != "KASIR" && role != "MEMBER" {
		return &ValidasiError{"Peran (role) tidak valid. Pilih ADMIN, KASIR, atau MEMBER"}
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	return s.Users.Tambah(ctx, nama, email, string(hash), role)
}

// Login mengembalikan token session baru jika email dan password cocok.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	id, hash, err := s.Users.HashByEmail(ctx, email)

	if errors.Is(err, model.ErrNotFound) {
		return "", model.ErrLoginGagal
	}

	if err != nil {
		return "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", model.ErrLoginGagal
	}

	return s.Sessions.Buat(ctx, id)
}

func (s *AuthService) UserDariToken(ctx context.Context, token string) (model.User, error) {
	return s.Sessions.UserDariToken(ctx, token)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.Sessions.Hapus(ctx, token)
}

