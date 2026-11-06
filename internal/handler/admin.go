package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type AdminHandler struct {
	Repo *repository.AdminRepo
	View *Renderer
}

func (h *AdminHandler) Cabang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.FormValue("kode") == "" || r.FormValue("nama") == "" || r.FormValue("alamat") == "" {
			http.Error(w, "Kode, nama, dan alamat cabang wajib diisi", http.StatusBadRequest)
			return
		}
		if e := h.Repo.TambahCabang(r.Context(), model.Cabang{Kode: r.FormValue("kode"), Nama: r.FormValue("nama"), Alamat: r.FormValue("alamat"), Telepon: r.FormValue("telepon")}); e != nil {
			gagal(w, e)
			return
		}
		http.Redirect(w, r, "/cabang", 303)
		return
	}
	x, e := h.Repo.Cabang(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	u, _ := middleware.UserDari(r.Context())
	h.View.Tampil(w, "cabang.html", struct {
		User   model.User
		Cabang []model.Cabang
	}{u, x})
}

func (h *AdminHandler) Member(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if r.FormValue("nomor") == "" || r.FormValue("nama") == "" || r.FormValue("no_hp") == "" {
			http.Error(w, "Nomor member, nama, dan nomor HP wajib diisi", http.StatusBadRequest)
			return
		}
		if e := h.Repo.TambahMember(r.Context(), model.Member{Nomor: r.FormValue("nomor"), Nama: r.FormValue("nama"), NoHP: r.FormValue("no_hp"), Email: r.FormValue("email"), Alamat: r.FormValue("alamat")}); e != nil {
			gagal(w, e)
			return
		}
		http.Redirect(w, r, "/member", 303)
		return
	}
	x, e := h.Repo.Member(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	u, _ := middleware.UserDari(r.Context())
	h.View.Tampil(w, "member.html", struct {
		User   model.User
		Member []model.Member
	}{u, x})
}

func (h *AdminHandler) Promo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		a, _ := time.Parse("2006-01-02T15:04", r.FormValue("mulai"))
		b, _ := time.Parse("2006-01-02T15:04", r.FormValue("berakhir"))
		if r.FormValue("kode") == "" || r.FormValue("nama") == "" || a.IsZero() || b.IsZero() || !b.After(a) || angka(r.FormValue("nilai")) < 0 || angka(r.FormValue("minimal")) < 0 {
			http.Error(w, "Data promo atau periode tidak valid", http.StatusBadRequest)
			return
		}
		x := model.Promo{Kode: r.FormValue("kode"), Nama: r.FormValue("nama"), Jenis: r.FormValue("jenis"), Nilai: angka(r.FormValue("nilai")), Minimal: angka(r.FormValue("minimal")), Maksimal: angka(r.FormValue("maksimal")), CabangID: angka(r.FormValue("cabang_id")), TargetProdukID: angka(r.FormValue("produk_id")), TargetKategoriID: angka(r.FormValue("kategori_id")), MulaiJam: r.FormValue("mulai_jam"), BerakhirJam: r.FormValue("berakhir_jam"), Mulai: a, Berakhir: b}
		if e := h.Repo.TambahPromo(r.Context(), x); e != nil {
			gagal(w, e)
			return
		}
		http.Redirect(w, r, "/promo", 303)
		return
	}
	x, e := h.Repo.Promo(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	p, e := h.Repo.Produk(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	k, e := h.Repo.Kategori(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	c, e := h.Repo.Cabang(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	u, _ := middleware.UserDari(r.Context())
	h.View.Tampil(w, "promo.html", struct {
		User     model.User
		Promo    []model.Promo
		Produk   []model.Produk
		Kategori []model.Kategori
		Cabang   []model.Cabang
	}{u, x, p, k, c})
}

func (h *AdminHandler) Voucher(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if angka(r.FormValue("promo_id")) <= 0 || r.FormValue("kode") == "" || angka(r.FormValue("kuota")) <= 0 {
			http.Error(w, "Promo, kode voucher, dan kuota wajib valid", http.StatusBadRequest)
			return
		}
		if e := h.Repo.TambahVoucher(r.Context(), model.Voucher{PromoID: angka(r.FormValue("promo_id")), Kode: r.FormValue("kode"), Kuota: angka(r.FormValue("kuota"))}); e != nil {
			gagal(w, e)
			return
		}
		http.Redirect(w, r, "/voucher", 303)
		return
	}
	x, e := h.Repo.Voucher(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	p, _ := h.Repo.Promo(r.Context())
	u, _ := middleware.UserDari(r.Context())
	h.View.Tampil(w, "voucher.html", struct {
		User    model.User
		Voucher []model.Voucher
		Promo   []model.Promo
	}{u, x, p})
}

func angka(s string) int { n, _ := strconv.Atoi(s); return n }
