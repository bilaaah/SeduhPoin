package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

// LaporanHandler melayani halaman Laporan Analitik beserta ekspor CSV-nya.
type LaporanHandler struct {
	Repo *repository.AdminRepo
	View *Renderer
}

// Laporan menampilkan ketiga laporan wajib Lingkup 6.
// Pembatasan peran dilakukan oleh middleware RequireRole pada route.
func (h *LaporanHandler) Laporan(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())

	data, err := h.Repo.LaporanAnalitik(r.Context(), parseFilter(r))
	if err != nil {
		gagal(w, err)
		return
	}
	data.User = u
	h.View.Tampil(w, "laporan.html", data)
}

// ExportCSV mengekspor salah satu laporan ke berkas CSV yang siap dibuka di Excel.
// Pembatasan peran dilakukan oleh middleware RequireRole pada route.
func (h *LaporanHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	data, err := h.Repo.LaporanAnalitik(r.Context(), parseFilter(r))
	if err != nil {
		gagal(w, err)
		return
	}

	jenis := r.URL.Query().Get("jenis")
	switch jenis {
	case "promo", "poin", "member":
	default:
		jenis = "promo"
	}

	nama := "laporan-" + jenis + "-" + time.Now().Format("20060102") + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+nama+"\"")
	// BOM UTF-8 supaya Excel di Indonesia membaca karakter dengan benar.
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	switch jenis {
	case "promo":
		tulisCSVPromo(cw, data)
	case "poin":
		tulisCSVPoin(cw, data)
	case "member":
		tulisCSVMember(cw, data)
	}
}

// parseFilter membaca parameter filter laporan dari query string.
func parseFilter(r *http.Request) model.FilterLaporan {
	var f model.FilterLaporan
	if v, err := strconv.Atoi(r.URL.Query().Get("cabang")); err == nil && v > 0 {
		f.CabangID = v
	}
	if p := r.URL.Query().Get("periode"); p != "" {
		if _, err := time.Parse("2006-01", p); err == nil {
			f.Periode = p
		}
	}
	return f
}

func tulisCSVPromo(cw *csv.Writer, d model.LaporanAnalitik) {
	cw.Write([]string{"LAPORAN BIAYA REALISASI SETIAP PROMO"})
	cw.Write([]string{"Tanggal Ekspor", time.Now().Format("02/01/2006 15:04")})
	cw.Write([]string{"Total Biaya Promo", strconv.Itoa(d.TotalBiayaPromo)})
	cw.Write([]string{"Total Pemakaian Promo", strconv.Itoa(d.TotalPemakaian)})
	cw.Write([]string{})
	cw.Write([]string{"Kode Promo", "Nama Promo", "Jenis Promo", "Jumlah Pemakaian", "Total Biaya Diskon (Rp)", "Porsi Biaya (%)"})
	for _, p := range d.ListPromo {
		cw.Write([]string{
			p.Kode,
			p.Nama,
			p.Jenis,
			strconv.Itoa(p.JumlahPemakaian),
			strconv.Itoa(p.TotalBiaya),
			fmt.Sprintf("%.2f", p.Persen),
		})
	}
	cw.Write([]string{"TOTAL", "", "", strconv.Itoa(d.TotalPemakaian), strconv.Itoa(d.TotalBiayaPromo), "100.00"})
}

func tulisCSVPoin(cw *csv.Writer, d model.LaporanAnalitik) {
	cw.Write([]string{"LAPORAN POIN YANG MASIH BERLAKU"})
	cw.Write([]string{"Tanggal Ekspor", time.Now().Format("02/01/2006 15:04")})
	cw.Write([]string{"Total Poin Aktif", strconv.Itoa(d.PoinAktif)})
	cw.Write([]string{"Nilai Rupiah Poin", strconv.Itoa(d.NilaiPoin)})
	cw.Write([]string{"Konversi", "1 Poin = Rp1.000"})
	cw.Write([]string{})
	cw.Write([]string{"RINCIAN PER MEMBER"})
	cw.Write([]string{"No. Member", "Nama Member", "Tingkat", "Jumlah Batch", "Total Poin", "Nilai Rupiah (Rp)", "Kedaluwarsa Terdekat"})
	for _, m := range d.PoinPerMember {
		cw.Write([]string{
			m.NomorMember,
			m.Nama,
			m.Tingkat,
			strconv.Itoa(m.JumlahBatch),
			strconv.Itoa(m.TotalPoin),
			strconv.Itoa(m.NilaiRupiah),
			m.KedaluarsaTerdekat,
		})
	}
	cw.Write([]string{"TOTAL", "", "", "", strconv.Itoa(d.PoinAktif), strconv.Itoa(d.NilaiPoin), ""})
	cw.Write([]string{})
	cw.Write([]string{"RINCIAN PER CABANG ASAL PEROLEHAN POIN"})
	cw.Write([]string{"Cabang", "Total Poin", "Nilai Rupiah (Rp)"})
	for _, c := range d.PoinPerCabang {
		cw.Write([]string{c.NamaCabang, strconv.Itoa(c.TotalPoin), strconv.Itoa(c.NilaiRupiah)})
	}
}

func tulisCSVMember(cw *csv.Writer, d model.LaporanAnalitik) {
	cw.Write([]string{"LAPORAN PERKEMBANGAN JUMLAH MEMBER PER TINGKAT"})
	cw.Write([]string{"Tanggal Ekspor", time.Now().Format("02/01/2006 15:04")})
	cw.Write([]string{"Total Member", strconv.Itoa(d.TotalMember)})
	cw.Write([]string{})
	cw.Write([]string{"DISTRIBUSI KEANGGOTAAN"})
	cw.Write([]string{"Tingkat", "Jumlah Member", "Porsi (%)"})
	for _, t := range d.Tier {
		cw.Write([]string{t.Tingkat, strconv.Itoa(t.Jumlah), fmt.Sprintf("%.2f", t.Persen)})
	}
	cw.Write([]string{"TOTAL", strconv.Itoa(d.TotalMember), "100.00"})
	cw.Write([]string{})
	cw.Write([]string{"PERTUMBUHAN MEMBER BARU PER BULAN"})
	cw.Write([]string{"Bulan", "Label Bulan", "Member Baru", "Total Kumulatif"})
	for _, p := range d.Pertumbuhan {
		cw.Write([]string{p.Bulan, p.LabelBulan, strconv.Itoa(p.JumlahBaru), strconv.Itoa(p.TotalKumulatif)})
	}
	cw.Write([]string{"TOTAL", "", strconv.Itoa(d.TotalMemberBaru), ""})
}
