package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/service"
)

type ctxKey struct{}

type Auth struct {
	Svc *service.AuthService
}

// RequireLogin memastikan pengguna sudah login sebelum mengakses halaman privat.
// Jika belum login, diarahkan ke /login.
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
			http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole membatasi akses suatu handler hanya untuk peran tertentu.
// Harus digunakan di dalam RequireLogin agar data user sudah ada di context.
// Jika peran tidak diizinkan, menampilkan halaman 403 Forbidden.
func (a *Auth) RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserDari(r.Context())
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			if allowed[u.Role] {
				next.ServeHTTP(w, r)
				return
			}

			// Peran tidak diizinkan → tampilkan halaman 403
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, halaman403(u.Nama, u.Role))
		})
	}
}

// halaman403 menghasilkan HTML sederhana bertema kopi untuk error 403.
func halaman403(nama, role string) string {
	label := map[string]string{
		"ADMIN":  "Admin / Pemilik",
		"KASIR":  "Kasir Operasional",
		"MEMBER": "Member Loyalitas",
	}
	roleLabel, ok := label[role]
	if !ok {
		roleLabel = role
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="id">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>403 – Akses Ditolak | SeduhPoin</title>
  <style>
    *{box-sizing:border-box;margin:0;padding:0}
    body{font-family:Inter,system-ui,sans-serif;background:#fffaf3;color:#2d211d;
         display:flex;align-items:center;justify-content:center;min-height:100vh;padding:24px}
    .card{background:#fff;border:1px solid #eee5dc;border-radius:20px;padding:48px 40px;
          max-width:480px;width:100%%;box-shadow:0 16px 40px rgba(65,44,32,.08);text-align:center}
    .icon{font-size:64px;margin-bottom:20px}
    h1{font:700 28px Georgia,serif;color:#2d211d;margin-bottom:12px}
    p{color:#8b817b;line-height:1.7;margin-bottom:8px;font-size:14px}
    .badge{display:inline-block;background:#fdf0e3;color:#c88952;border:1px solid #f0d9bb;
           border-radius:20px;padding:4px 14px;font-size:12px;font-weight:700;margin-bottom:24px}
    .btn{display:inline-block;background:#68483a;color:#fff;padding:12px 28px;
         border-radius:10px;font-weight:700;text-decoration:none;font-size:14px}
    .btn:hover{background:#2d211d}
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">🚫</div>
    <h1>Akses Ditolak</h1>
    <div class="badge">%s — %s</div>
    <p>Halaman ini tidak dapat diakses oleh peran Anda.</p>
    <p>Silakan kembali ke beranda atau hubungi administrator jika Anda membutuhkan akses ini.</p>
    <br>
    <a href="/" class="btn">☕ Kembali ke Beranda</a>
  </div>
</body>
</html>`, nama, roleLabel)
}

// UserDari mengambil data User dari context request.
func UserDari(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(ctxKey{}).(model.User)
	return u, ok
}

