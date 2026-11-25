package handler

import (
	"net/http"
	"strconv"
)

// BatalTransaksi menangani pembatalan transaksi oleh Kasir / Admin (Lingkup 4 - Hilal)
func (h *HalamanHandler) BatalTransaksi(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	transaksiID, err := strconv.Atoi(idStr)
	if err != nil || transaksiID <= 0 {
		http.Error(w, "ID transaksi tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Repo.BatalkanTransaksi(r.Context(), transaksiID)
	if err != nil {
		http.Error(w, "Gagal membatalkan transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/kasir/riwayat", http.StatusSeeOther)
}
