package handler

import (
	"net/http"
	"strconv"
)

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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}