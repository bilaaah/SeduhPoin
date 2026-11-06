package handler

import (
	"errors"
	"net/http"

	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/service"
)

type AuthForm struct {
	Nama  string
	Email string
	Role  string
	Error string
}

type AuthHandler struct {
	Svc          *service.AuthService
	View         *Renderer
	CookieSecure bool
}

func (h *AuthHandler) FormRegister(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "auth/register.html", AuthForm{Role: "ADMIN"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	f := AuthForm{
		Nama:  r.PostFormValue("nama"),
		Email: r.PostFormValue("email"),
		Role:  r.PostFormValue("role"),
	}

	err := h.Svc.Register(
		r.Context(),
		f.Nama,
		f.Email,
		r.PostFormValue("password"),
		f.Role,
	)

	var ve *service.ValidasiError

	switch {
	case err == nil:
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return

	case errors.As(err, &ve):
		f.Error = ve.Pesan

	case errors.Is(err, model.ErrEmailTerpakai):
		f.Error = "Email sudah terdaftar"

	default:
		gagal(w, err)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	h.View.Tampil(w, "auth/register.html", f)
}

func (h *AuthHandler) FormLogin(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "auth/login.html", AuthForm{})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")

	token, err := h.Svc.Login(
		r.Context(),
		email,
		r.PostFormValue("password"),
	)

	if errors.Is(err, model.ErrLoginGagal) {
		w.WriteHeader(http.StatusUnauthorized)

		h.View.Tampil(w, "auth/login.html", AuthForm{
			Email: email,
			Error: "Email atau password salah",
		})

		return
	}

	if err != nil {
		gagal(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 hari
		HttpOnly: true,             // tidak bisa dibaca JavaScript
		Secure:   h.CookieSecure,   // true di production (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		h.Svc.Logout(r.Context(), c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

