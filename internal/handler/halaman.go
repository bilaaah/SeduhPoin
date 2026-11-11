package handler

import (
	"errors"
	"net/http"
	"net/url"
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
	ringkasan.User = u
	h.View.Tampil(w, "home.html", ringkasan)
}

// TransaksiKasir menangani proses penjualan oleh Kasir
func (h *HalamanHandler) SimpanTransaksi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Data transaksi tidak valid", http.StatusBadRequest)
			return
		}

		cabangID, _ := strconv.Atoi(r.FormValue("cabang_id"))
		memberID, _ := strconv.Atoi(r.FormValue("member_id"))
		subtotal, _ := strconv.Atoi(r.FormValue("subtotal"))
		poinTukar, _ := strconv.Atoi(r.FormValue("poin_ditukar"))
		poinDapat, _ := strconv.Atoi(r.FormValue("poin_diperoleh"))

		if cabangID <= 0 || subtotal <= 0 {
			http.Error(w, "Data transaksi tidak valid", http.StatusBadRequest)
			return
		}

		in := model.TransaksiInput{
			CabangID:      cabangID,
			MemberID:      memberID,
			PoinDigunakan: poinTukar,
			Subtotal:      subtotal,
			PoinDiperoleh: poinDapat,
			Promo:         bacaAlokasiPromo(r.Form),
		}

		if err := h.Repo.TambahTransaksiKasir(r.Context(), in); err != nil {
			var validasi repository.ErrValidasi
			if errors.As(err, &validasi) {
				http.Error(w, validasi.Error(), http.StatusBadRequest)
				return
			}
			gagal(w, err)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// bacaAlokasiPromo membaca promo yang diterapkan kasir. Nama field memakai
// array implisit: form["promo_id"], form["promo_diskon"], dan form["promo_voucher"]
// berisi nilai urut sesuai baris form.
func bacaAlokasiPromo(form url.Values) []model.AlokasiPromo {
	ids := form["promo_id"]
	diskons := form["promo_diskon"]
	vouchers := form["promo_voucher"]

	var out []model.AlokasiPromo
	for i, v := range ids {
		promoID, err := strconv.Atoi(v)
		if err != nil || promoID <= 0 {
			continue
		}
		diskon := 0
		if i < len(diskons) {
			diskon, _ = strconv.Atoi(diskons[i])
		}
		kode := ""
		if i < len(vouchers) {
			kode = strings.TrimSpace(vouchers[i])
		}
		out = append(out, model.AlokasiPromo{PromoID: promoID, VoucherKode: kode, Diskon: diskon})
	}
	return out
}

// BatalTransaksiKasir menangani pembatalan transaksi oleh Kasir
func (h *HalamanHandler) BatalTransaksi(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id > 0 {
		_ = h.Repo.BatalkanTransaksi(r.Context(), id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
