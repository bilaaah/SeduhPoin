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

	// privat: semua route yang wajib login
	privat := http.NewServeMux()
	privat.HandleFunc("GET /{$}", halaman.Home)

	// Lingkup 1 (Nabila)
	privat.HandleFunc("/cabang", admin.Cabang)
	privat.HandleFunc("GET /produk", admin.Produk)
	privat.HandleFunc("POST /produk", admin.Produk)
	privat.HandleFunc("POST /produk/{id}/hapus", admin.HapusProduk)
	privat.HandleFunc("/member", admin.Member)

	// Lingkup 2 (Nabila)
	privat.HandleFunc("/promo", admin.Promo)
	privat.HandleFunc("/voucher", admin.Voucher)

	// Laporan (Admin/Pemilik)
	privat.HandleFunc("GET /laporan", halaman.Laporan)

	// Transaksi & Batal (Kasir)
	privat.HandleFunc("POST /transaksi", halaman.SimpanTransaksi)
	privat.HandleFunc("POST /transaksi/{id}/batal", halaman.BatalTransaksi)

	// Logout
	privat.HandleFunc("POST /logout", authH.Logout)

	// publik: boleh diakses tanpa login
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", authH.FormLogin)
	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("GET /register", authH.FormRegister)
	mux.HandleFunc("POST /register", authH.Register)

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
