package handler

import (
	"net/http"
	"strconv"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type HalamanHandler struct {
	Repo *repository.AdminRepo
	View *Renderer
}

func (h *HalamanHandler) Home(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())

	// 1. Jika KASIR yang login
	if u.Role == "KASIR" {
		cabang, _ := h.Repo.Cabang(r.Context())
		produk, _ := h.Repo.Produk(r.Context())
		member, _ := h.Repo.Member(r.Context())
		promo, _ := h.Repo.Promo(r.Context())
		voucher, _ := h.Repo.Voucher(r.Context())
		trx, _ := h.Repo.SemuaTransaksi(r.Context())

		data := model.RingkasanData{
			User: u,
			KasirData: model.KasirData{
				User:      u,
				Cabang:    cabang,
				Produk:    produk,
				Member:    member,
				Promo:     promo,
				Voucher:   voucher,
				Transaksi: trx,
			},
		}
		h.View.Tampil(w, "home.html", data)
		return
	}

	// 2. Jika MEMBER yang login
	if u.Role == "MEMBER" {
		mbr, _ := h.Repo.MemberByEmail(r.Context(), u.Email)
		poinAktif, batchPoin, _ := h.Repo.PoinByMemberID(r.Context(), mbr.ID)
		riwayat, _ := h.Repo.TransaksiByMemberID(r.Context(), mbr.ID)

		sisaGold := 1500000 - mbr.Total
		if sisaGold < 0 {
			sisaGold = 0
		}

		data := model.RingkasanData{
			User: u,
			MemberData: model.MemberPortalData{
				User:           u,
				Member:         mbr,
				PoinAktif:      poinAktif,
				NilaiRupiah:    poinAktif * 1000,
				SisaMenujuGold: sisaGold,
				BatchPoin:      batchPoin,
				Riwayat:        riwayat,
			},
		}
		h.View.Tampil(w, "home.html", data)
		return
	}

	// 3. Default: ADMIN / PEMILIK
	ringkasan, err := h.Repo.Ringkasan(r.Context())
	if err != nil {
		ringkasan.User = u
		h.View.Tampil(w, "home.html", ringkasan)
		return
	}
	laporan, _ := h.Repo.Laporan(r.Context())
	ringkasan.User = u
	ringkasan.LaporanData = laporan
	h.View.Tampil(w, "home.html", ringkasan)
}

// Laporan khusus Admin / Pemilik
func (h *HalamanHandler) Laporan(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	lap, err := h.Repo.Laporan(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	lap.User = u
	h.View.Tampil(w, "laporan.html", lap)
}

// TransaksiKasir menangani proses penjualan oleh Kasir
func (h *HalamanHandler) SimpanTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cabangID, _ := strconv.Atoi(r.FormValue("cabang_id"))
		memberID, _ := strconv.Atoi(r.FormValue("member_id"))
		subtotal, _ := strconv.Atoi(r.FormValue("subtotal"))
		diskon, _ := strconv.Atoi(r.FormValue("total_diskon"))
		bayar, _ := strconv.Atoi(r.FormValue("total_bayar"))
		poinTukar, _ := strconv.Atoi(r.FormValue("poin_ditukar"))
		poinDapat, _ := strconv.Atoi(r.FormValue("poin_diperoleh"))

		if cabangID <= 0 || subtotal <= 0 {
			http.Error(w, "Data transaksi tidak valid", http.StatusBadRequest)
			return
		}

		if err := h.Repo.TambahTransaksiKasir(r.Context(), cabangID, memberID, subtotal, diskon, bayar, poinTukar, poinDapat); err != nil {
			gagal(w, err)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// BatalTransaksiKasir menangani pembatalan transaksi oleh Kasir
func (h *HalamanHandler) BatalTransaksi(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id > 0 {
		_ = h.Repo.BatalkanTransaksi(r.Context(), id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
