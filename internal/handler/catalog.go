package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
)

type ProdukPage struct {
	User     model.User
	Produk   []model.Produk
	Kategori []model.Kategori
	Edit     model.Produk
	Mode     string
}

func (h *AdminHandler) Produk(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		x := produkDariForm(r)
		if x.Nama == "" || x.KategoriID <= 0 || x.Harga < 0 || x.Stok < 0 {
			http.Error(w, "Nama, kategori, harga, dan stok produk wajib valid", http.StatusBadRequest)
			return
		}
		var err error
		if x.ID > 0 {
			err = h.Repo.UbahProduk(r.Context(), x)
		} else {
			err = h.Repo.TambahProduk(r.Context(), x)
		}
		if err != nil {
			gagal(w, err)
			return
		}
		http.Redirect(w, r, "/produk", http.StatusSeeOther)
		return
	}
	items, err := h.Repo.Produk(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	kategori, err := h.Repo.Kategori(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	u, _ := middleware.UserDari(r.Context())
	data := ProdukPage{User: u, Produk: items, Kategori: kategori, Mode: "list"}
	if r.URL.Query().Get("action") == "tambah" {
		data.Mode = "tambah"
	} else if id, _ := strconv.Atoi(r.URL.Query().Get("edit")); id > 0 {
		for _, x := range items {
			if x.ID == id {
				data.Edit = x
				data.Mode = "edit"
				break
			}
		}
	}
	h.View.Tampil(w, "produk.html", data)
}

func (h *AdminHandler) HapusProduk(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID produk tidak valid", 400)
		return
	}
	if err := h.Repo.HapusProduk(r.Context(), id); err != nil {
		gagal(w, err)
		return
	}
	http.Redirect(w, r, "/produk", http.StatusSeeOther)
}

func produkDariForm(r *http.Request) model.Produk {
	id, _ := strconv.Atoi(r.FormValue("id"))
	kategori, _ := strconv.Atoi(r.FormValue("kategori_id"))
	harga, _ := strconv.Atoi(r.FormValue("harga"))
	stok, _ := strconv.Atoi(r.FormValue("stok"))
	return model.Produk{ID: id, KategoriID: kategori, Nama: strings.TrimSpace(r.FormValue("nama")), Harga: harga, Stok: stok, Keterangan: strings.TrimSpace(r.FormValue("keterangan"))}
}
