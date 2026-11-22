package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}

	cabangID, err := strconv.Atoi(r.FormValue("cabang_id"))
	if err != nil || cabangID <= 0 {
		http.Error(w, "Cabang tidak valid", http.StatusBadRequest)
		return
	}

	memberID, _ := strconv.Atoi(r.FormValue("member_id"))
	voucherID, _ := strconv.Atoi(r.FormValue("voucher_id"))
	diskon, _ := strconv.Atoi(r.FormValue("total_diskon"))
	bayar, _ := strconv.Atoi(r.FormValue("total_bayar"))
	poinTukar, _ := strconv.Atoi(r.FormValue("poin_ditukar"))
	poinDapat, _ := strconv.Atoi(r.FormValue("poin_diperoleh"))

	// Membaca seluruh jumlah produk dari form
	var items []model.TransaksiItemInput

	for key, values := range r.PostForm {
		if !strings.HasPrefix(key, "jumlah_produk_") || len(values) == 0 {
			continue
		}

		produkID, err := strconv.Atoi(strings.TrimPrefix(key, "jumlah_produk_"))
		if err != nil || produkID <= 0 {
			http.Error(w, "ID produk tidak valid", http.StatusBadRequest)
			return
		}

		jumlah, err := strconv.Atoi(values[0])
		if err != nil || jumlah < 0 {
			http.Error(w, "Jumlah produk tidak valid", http.StatusBadRequest)
			return
		}

		if jumlah > 0 {
			items = append(items, model.TransaksiItemInput{
				ProdukID: produkID,
				Jumlah:   jumlah,
			})
		}
	}

	if len(items) == 0 {
		http.Error(w, "Pilih minimal satu produk", http.StatusBadRequest)
		return
	}

	// Penyimpanan transaksi akan menggunakan daftar produk.
	// Repository akan memvalidasi harga dan stok berdasarkan database.
	if err := h.Repo.TambahTransaksiKasir(
		r.Context(),
		cabangID,
		memberID,
		voucherID,
		diskon,
		bayar,
		poinTukar,
		poinDapat,
		items,
	); err != nil {
		gagal(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}


