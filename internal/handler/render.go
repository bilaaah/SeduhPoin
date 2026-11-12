package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

// rupiah memformat angka jadi "Rp39.000.000"
func rupiah(n int) string {
	s := strconv.Itoa(n)
	hasil := ""
	for i, c := range s {
		// sisipkan titik tiap 3 digit dari kanan, kecuali setelah tanda minus
		if i > 0 && (len(s)-i)%3 == 0 && s[i-1] != '-' {
			hasil += "."
		}
		hasil += string(c)
	}
	return "Rp" + hasil
}

func addOne(n int) int {
	return n + 1
}

func mul(a, b int) int {
	return a * b
}

// persen mengubah angka persen menjadi teks, misal 42.9 menjadi "42.9%".
func persen(f float64) string {
	return strconv.FormatFloat(f, 'f', 1, 64) + "%"
}

// desimal mengubah angka menjadi teks desimal satu digit, dipakai untuk
// nilai persentase di dalam atribut style.
func desimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 1, 64)
}

func userRole(data any) string {
	if data == nil {
		return ""
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName("User")
		if field.IsValid() {
			if u, ok := field.Interface().(model.User); ok {
				return u.Role
			}
		}
		if val.Type() == reflect.TypeOf(model.User{}) {
			return val.Interface().(model.User).Role
		}
	}
	return ""
}

func userNama(data any) string {
	if data == nil {
		return ""
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName("User")
		if field.IsValid() {
			if u, ok := field.Interface().(model.User); ok {
				return u.Nama
			}
		}
		if val.Type() == reflect.TypeOf(model.User{}) {
			return val.Interface().(model.User).Nama
		}
	}
	return ""
}

var funcs = template.FuncMap{
	"rupiah":   rupiah,
	"add_one":  addOne,
	"mul":      mul,
	"persen":   persen,
	"desimal":  desimal,
	"userRole": userRole,
	"userNama": userNama,
}

type Renderer struct {
	tpl map[string]*template.Template
}

func NewRenderer() *Renderer {
	halaman := []string{
		"home.html",
		"produk.html",
		"cabang.html",
		"member.html",
		"promo.html",
		"voucher.html",
		"laporan.html",
		"kasir_transaksi.html",
		"kasir_riwayat.html",
		"member_poin.html",
		"member_riwayat.html",
		"auth/login.html",
		"auth/register.html",
	}

	base := "templates"
	for _, candidate := range []string{"templates", filepath.Join("..", "templates"), filepath.Join("..", "..", "templates")} {
		if _, err := os.Stat(filepath.Join(candidate, "layout.html")); err == nil {
			base = candidate
			break
		}
	}

	rd := &Renderer{tpl: make(map[string]*template.Template)}
	for _, file := range halaman {
		targetPath := filepath.Join(base, filepath.FromSlash(file))
		rd.tpl[file] = template.Must(template.New("layout").Funcs(funcs).ParseFiles(
			filepath.Join(base, "layout.html"),
			filepath.Join(base, "header.html"),
			filepath.Join(base, "sidebar.html"),
			filepath.Join(base, "footer.html"),
			targetPath,
		))
	}
	return rd
}

// Tampil me-render ke buffer dulu supaya error template tidak menghasilkan halaman setengah jadi.
func (rd *Renderer) Tampil(w http.ResponseWriter, file string, data any) {
	var buf bytes.Buffer
	t, ok := rd.tpl[file]
	if !ok {
		log.Printf("template tidak terdaftar: %s", file)
		http.Error(w, "Halaman tidak ditemukan", http.StatusNotFound)
		return
	}
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Printf("render %s: %v", file, err)
		http.Error(w, "Template SeduhPoin bermasalah: "+err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}
