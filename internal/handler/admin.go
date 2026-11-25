package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type AdminHandler struct {
	Repo *repository.AdminRepo
	View *Renderer
}

type CabangPage struct {
	User   model.User
	Cabang []model.Cabang
	Edit   model.Cabang
	Mode   string
}

func (h *AdminHandler) Cabang(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		nama := strings.TrimSpace(r.FormValue("nama"))
		alamat := strings.TrimSpace(r.FormValue("alamat"))
		telepon := strings.TrimSpace(r.FormValue("telepon"))
		aktif := r.FormValue("aktif") == "1" || r.FormValue("aktif") == "on" || (id == 0 && r.FormValue("aktif") != "0")

		if id > 0 {
			if nama == "" || alamat == "" {
				http.Error(w, "Nama dan alamat cabang wajib diisi", http.StatusBadRequest)
				return
			}
			if err := h.Repo.UbahCabang(r.Context(), model.Cabang{
				ID:      id,
				Nama:    nama,
				Alamat:  alamat,
				Telepon: telepon,
				Aktif:   aktif,
			}); err != nil {
				gagal(w, err)
				return
			}
		} else {
			kode := strings.ToUpper(strings.TrimSpace(r.FormValue("kode")))
			if kode == "" || nama == "" || alamat == "" {
				http.Error(w, "Kode, nama, dan alamat cabang wajib diisi", http.StatusBadRequest)
				return
			}
			if err := h.Repo.TambahCabang(r.Context(), model.Cabang{
				Kode:    kode,
				Nama:    nama,
				Alamat:  alamat,
				Telepon: telepon,
				Aktif:   true,
			}); err != nil {
				gagal(w, err)
				return
			}
		}
		http.Redirect(w, r, "/cabang", http.StatusSeeOther)
		return
	}

	list, err := h.Repo.Cabang(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	u, _ := middleware.UserDari(r.Context())

	page := CabangPage{User: u, Cabang: list}
	if editID, _ := strconv.Atoi(r.URL.Query().Get("edit")); editID > 0 {
		for _, c := range list {
			if c.ID == editID {
				page.Edit = c
				page.Mode = "edit"
				break
			}
		}
	}
	h.View.Tampil(w, "cabang.html", page)
}

func (h *AdminHandler) ToggleCabang(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id > 0 {
		_ = h.Repo.ToggleCabang(r.Context(), id)
	}
	http.Redirect(w, r, "/cabang", http.StatusSeeOther)
}

type MemberPage struct {
	User   model.User
	Member []model.Member
	Edit   model.Member
	Mode   string
}

func (h *AdminHandler) Member(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		nama := strings.TrimSpace(r.FormValue("nama"))
		noHP := strings.TrimSpace(r.FormValue("no_hp"))
		email := strings.TrimSpace(r.FormValue("email"))
		alamat := strings.TrimSpace(r.FormValue("alamat"))
		aktif := r.FormValue("aktif") == "1" || r.FormValue("aktif") == "on" || (id == 0 && r.FormValue("aktif") != "0")

		if id > 0 {
			if nama == "" || noHP == "" {
				http.Error(w, "Nama dan nomor HP member wajib diisi", http.StatusBadRequest)
				return
			}
			if err := h.Repo.UbahMember(r.Context(), model.Member{
				ID:     id,
				Nama:   nama,
				NoHP:   noHP,
				Email:  email,
				Alamat: alamat,
				Aktif:  aktif,
			}); err != nil {
				gagal(w, err)
				return
			}
		} else {
			nomor := strings.ToUpper(strings.TrimSpace(r.FormValue("nomor")))
			if nomor == "" || nama == "" || noHP == "" {
				http.Error(w, "Nomor member, nama, dan nomor HP wajib diisi", http.StatusBadRequest)
				return
			}
			if err := h.Repo.TambahMember(r.Context(), model.Member{
				Nomor:  nomor,
				Nama:   nama,
				NoHP:   noHP,
				Email:  email,
				Alamat: alamat,
			}); err != nil {
				gagal(w, err)
				return
			}
		}
		http.Redirect(w, r, "/member", http.StatusSeeOther)
		return
	}

	list, err := h.Repo.Member(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	u, _ := middleware.UserDari(r.Context())

	page := MemberPage{User: u, Member: list}
	if editID, _ := strconv.Atoi(r.URL.Query().Get("edit")); editID > 0 {
		for _, m := range list {
			if m.ID == editID {
				page.Edit = m
				page.Mode = "edit"
				break
			}
		}
	}
	h.View.Tampil(w, "member.html", page)
}

func (h *AdminHandler) ToggleMember(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if id > 0 {
		_ = h.Repo.ToggleMember(r.Context(), id)
	}
	http.Redirect(w, r, "/member", http.StatusSeeOther)
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
