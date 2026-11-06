package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/service"
)

type ctxKey struct{}

type Auth struct {
	Svc *service.AuthService
}

func (a *Auth) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		u, err := a.Svc.UserDariToken(r.Context(), c.Value)

		if errors.Is(err, model.ErrNotFound) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if err != nil {
			log.Printf("cek session: %v", err)
			http.Error(
				w,
				"Terjadi kesalahan pada server",
				http.StatusInternalServerError,
			)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserDari(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(ctxKey{}).(model.User)
	return u, ok
}

