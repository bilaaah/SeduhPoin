package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/bilaaah/SeduhPoin/internal/config"
	"github.com/bilaaah/SeduhPoin/internal/handler"
	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/repository"
	"github.com/bilaaah/SeduhPoin/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := bukaDB(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	view := handler.NewRenderer()
	adminRepo := &repository.AdminRepo{DB: db}
	halaman := &handler.HalamanHandler{Repo: adminRepo, View: view}
	admin := &handler.AdminHandler{Repo: adminRepo, View: view}
	laporan := &handler.LaporanHandler{Repo: adminRepo, View: view}
	tier := &handler.MemberTierHandler{Repo: &repository.MemberTierRepo{DB: db}, View: view}

	authSvc := &service.AuthService{
		Users:    &repository.UserRepo{DB: db},
		Sessions: &repository.SessionRepo{DB: db},
	}

	authH := &handler.AuthHandler{
		Svc:          authSvc,
		View:         view,
		CookieSecure: cfg.CookieSecure,
	}

	authMw := &middleware.Auth{Svc: authSvc}

	// Helper shortcut pembatasan peran (harus dipanggil setelah RequireLogin)
	adminOnly := authMw.RequireRole("ADMIN")
	adminKasir := authMw.RequireRole("ADMIN", "KASIR")
	memberOnly := authMw.RequireRole("ADMIN", "MEMBER")

	// privat: semua route yang wajib login
	privat := http.NewServeMux()

	// Dashboard — terbuka untuk semua peran (isi berbeda per peran)
	privat.HandleFunc("GET /{$}", halaman.Home)

	// ── Lingkup 1: Master Data ── hanya ADMIN (nabila)
	privat.Handle("GET /produk", adminOnly(http.HandlerFunc(admin.Produk)))
	privat.Handle("POST /produk", adminOnly(http.HandlerFunc(admin.Produk)))
	privat.Handle("POST /produk/{id}/hapus", adminOnly(http.HandlerFunc(admin.HapusProduk)))
	privat.Handle("GET /cabang", adminOnly(http.HandlerFunc(admin.Cabang)))
	privat.Handle("POST /cabang", adminOnly(http.HandlerFunc(admin.Cabang)))
	privat.Handle("POST /cabang/{id}/toggle", adminOnly(http.HandlerFunc(admin.ToggleCabang)))
	privat.Handle("GET /member", adminOnly(http.HandlerFunc(admin.Member)))
	privat.Handle("POST /member", adminOnly(http.HandlerFunc(admin.Member)))
	privat.Handle("POST /member/{id}/toggle", adminOnly(http.HandlerFunc(admin.ToggleMember)))

	// ── Lingkup 2: Promo & Voucher ── hanya ADMIN (nabila)
	privat.Handle("/promo", adminOnly(http.HandlerFunc(admin.Promo)))
	privat.Handle("POST /promo/{id}/toggle", adminOnly(http.HandlerFunc(admin.TogglePromo)))
	privat.Handle("/voucher", adminOnly(http.HandlerFunc(admin.Voucher)))
	privat.Handle("POST /voucher/{id}/hapus", adminOnly(http.HandlerFunc(admin.HapusVoucher)))

	// ── Lingkup 5: Poin & Tingkat Member (Dinda) ──
	privat.HandleFunc("GET /member/tier", tier.Halaman)
	privat.Handle("POST /member/tier/evaluasi", adminOnly(http.HandlerFunc(tier.Evaluasi)))

	// ── Lingkup 6: Laporan Analitik (Luhur) ── hanya ADMIN
	privat.Handle("GET /laporan", adminOnly(http.HandlerFunc(laporan.Laporan)))
	privat.Handle("GET /laporan/export.csv", adminOnly(http.HandlerFunc(laporan.ExportCSV)))

	// ── Operasional Kasir (Terminal POS & Pembatalan) ── ADMIN dan KASIR
	privat.Handle("GET /kasir/transaksi", adminKasir(http.HandlerFunc(halaman.KasirTransaksi)))
	privat.Handle("GET /kasir/riwayat", adminKasir(http.HandlerFunc(halaman.KasirRiwayat)))
	privat.Handle("POST /transaksi", adminKasir(http.HandlerFunc(halaman.SimpanTransaksi)))
	privat.Handle("POST /transaksi/{id}/batal", adminKasir(http.HandlerFunc(halaman.BatalTransaksi)))

	// ── Area Member Loyalitas ── ADMIN dan MEMBER
	privat.Handle("GET /member/poin", memberOnly(http.HandlerFunc(halaman.MemberPoin)))
	privat.Handle("GET /member/riwayat", memberOnly(http.HandlerFunc(halaman.MemberRiwayat)))

	// Logout — semua peran
	privat.HandleFunc("POST /logout", authH.Logout)

	// publik: boleh diakses tanpa login
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", authH.FormLogin)
	mux.HandleFunc("POST /login", authH.Login)
	// Registrasi dinonaktifkan: akun dikelola terpusat oleh Admin/Kedai
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.Handle("/", authMw.RequireLogin(privat))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           middleware.Logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server jalan di %s", cfg.Addr)
	log.Fatal(srv.ListenAndServe())
}

func bukaDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	log.Println("database terhubung")
	return db, nil
}
