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

func (h *HalamanHandler) ambilKasirData(r *http.Request, u model.User) model.KasirData {
	cabang, _ := h.Repo.Cabang(r.Context())
	produk, _ := h.Repo.Produk(r.Context())
	member, _ := h.Repo.Member(r.Context())
	promo, _ := h.Repo.Promo(r.Context())
	voucher, _ := h.Repo.Voucher(r.Context())
	trx, _ := h.Repo.SemuaTransaksi(r.Context())

	return model.KasirData{
		User:      u,
		Cabang:    cabang,
		Produk:    produk,
		Member:    member,
		Promo:     promo,
		Voucher:   voucher,
		Transaksi: trx,
	}
}

func (h *HalamanHandler) ambilMemberData(r *http.Request, u model.User) model.MemberPortalData {
	mbr, _ := h.Repo.MemberByEmail(r.Context(), u.Email)
	poinAktif, batchPoin, _ := h.Repo.PoinByMemberID(r.Context(), mbr.ID)
	riwayat, _ := h.Repo.TransaksiByMemberID(r.Context(), mbr.ID)

	sisaGold := 1500000 - mbr.Total
	if sisaGold < 0 {
		sisaGold = 0
	}

	return model.MemberPortalData{
		User:           u,
		Member:         mbr,
		PoinAktif:      poinAktif,
		NilaiRupiah:    poinAktif * 1000,
		SisaMenujuGold: sisaGold,
		BatchPoin:      batchPoin,
		Riwayat:        riwayat,
	}
}

func (h *HalamanHandler) Home(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())

	// 1. Jika KASIR yang login
	if u.Role == "KASIR" {
		kd := h.ambilKasirData(r, u)
		data := model.RingkasanData{
			User:      u,
			KasirData: kd,
		}
		h.View.Tampil(w, "home.html", data)
		return
	}

	// 2. Jika MEMBER yang login
	if u.Role == "MEMBER" {
		md := h.ambilMemberData(r, u)
		data := model.RingkasanData{
			User:       u,
			MemberData: md,
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

// KasirTransaksi menampilkan terminal khusus transaksi POS
func (h *HalamanHandler) KasirTransaksi(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	kd := h.ambilKasirData(r, u)
	h.View.Tampil(w, "kasir_transaksi.html", kd)
}

// KasirRiwayat menampilkan riwayat lengkap dan pembatalan pesanan kasir
func (h *HalamanHandler) KasirRiwayat(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	kd := h.ambilKasirData(r, u)
	h.View.Tampil(w, "kasir_riwayat.html", kd)
}

// MemberPoin menampilkan detail saldo poin aktif & riwayat masa berlaku FEFO 6 bulan
func (h *HalamanHandler) MemberPoin(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	md := h.ambilMemberData(r, u)
	h.View.Tampil(w, "member_poin.html", md)
}

// MemberRiwayat menampilkan histori lengkap belanja member di seluruh cabang
func (h *HalamanHandler) MemberRiwayat(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	md := h.ambilMemberData(r, u)
	h.View.Tampil(w, "member_riwayat.html", md)
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
		kodeVoucher := strings.TrimSpace(r.FormValue("kode_voucher"))

		items := bacaItemBelanja(r.Form)

		if cabangID <= 0 || (len(items) == 0 && subtotal <= 0) {
			http.Error(w, "Data transaksi tidak valid: pilih cabang dan minimal 1 produk", http.StatusBadRequest)
			return
		}

		in := model.TransaksiInput{
			CabangID:      cabangID,
			MemberID:      memberID,
			VoucherKode:   kodeVoucher,
			PoinDigunakan: poinTukar,
			Subtotal:      subtotal,
			PoinDiperoleh: poinDapat,
			Items:         items,
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
	http.Redirect(w, r, "/kasir/riwayat", http.StatusSeeOther)
}

func bacaItemBelanja(form url.Values) []model.ItemBelanja {
	pids := form["produk_id"]
	qtys := form["qty"]
	var out []model.ItemBelanja
	for i, v := range pids {
		pid, err := strconv.Atoi(v)
		if err != nil || pid <= 0 {
			continue
		}
		qty := 1
		if i < len(qtys) {
			qty, _ = strconv.Atoi(qtys[i])
		}
		if qty > 0 {
			out = append(out, model.ItemBelanja{ProdukID: pid, Qty: qty})
		}
	}
	return out
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
