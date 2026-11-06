package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

func gagal(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrNotFound) {
		http.Error(w, "Data tidak ditemukan", http.StatusNotFound)
		return
	}
	log.Printf("error: %v", err)
	// Aplikasi ini dipakai di lingkungan lokal XAMPP; tampilkan penyebabnya
	// agar kesalahan koneksi/schema dapat langsung diperbaiki.
	http.Error(w, "SeduhPoin mengalami kendala: "+err.Error(), http.StatusInternalServerError)
}
