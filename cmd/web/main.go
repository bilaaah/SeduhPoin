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
	halaman := &handler.HalamanHandler{View: view}

	admin := &handler.AdminHandler{Repo: &repository.AdminRepo{DB: db}, View: view}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", halaman.Home)
	mux.HandleFunc("GET /profil", halaman.Profil)

	mux.HandleFunc("/cabang", admin.Cabang)
	mux.HandleFunc("GET /produk", admin.Produk)
	mux.HandleFunc("POST /produk", admin.Produk)
	mux.HandleFunc("POST /produk/{id}/hapus", admin.HapusProduk)
	mux.HandleFunc("/member", admin.Member)
	mux.HandleFunc("/promo", admin.Promo)
	mux.HandleFunc("/voucher", admin.Voucher)

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
