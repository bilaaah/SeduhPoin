package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

const (
	// nilaiPoinPerPoin adalah konversi 1 poin menjadi rupiah saat penukaran.
	nilaiPoinPerPoin = 1000
	// jumlahBulanGrafik adalah panjang sumbu waktu grafik pertumbuhan member.
	jumlahBulanGrafik = 12
	// batasAwal dan batasAkhir tanpa batas dipakai saat filter periode kosong.
	batasAwal = "1970-01-01 00:00:00"
	batasAkhir = "2038-01-01 00:00:00"
)

// LaporanAnalitik menyusun ketiga laporan wajib Lingkup 6 dalam satu kali pemanggilan.
func (r *AdminRepo) LaporanAnalitik(ctx context.Context, f model.FilterLaporan) (model.LaporanAnalitik, error) {
	out := model.LaporanAnalitik{Filter: f}

	cabang, err := r.Cabang(ctx)
	if err != nil {
		return out, fmt.Errorf("daftar cabang laporan: %w", err)
	}
	out.Cabang = cabang

	periode, err := r.PeriodeTersedia(ctx)
	if err != nil {
		return out, fmt.Errorf("daftar periode laporan: %w", err)
	}
	out.PeriodeTersedia = periode

	promo, err := r.BiayaPromo(ctx, f)
	if err != nil {
		return out, fmt.Errorf("laporan biaya promo: %w", err)
	}
	out.ListPromo = promo
	for _, p := range promo {
		out.TotalBiayaPromo += p.TotalBiaya
		out.TotalPemakaian += p.JumlahPemakaian
	}
	for i := range out.ListPromo {
		if out.TotalBiayaPromo > 0 {
			out.ListPromo[i].Persen = float64(out.ListPromo[i].TotalBiaya) / float64(out.TotalBiayaPromo) * 100
		}
	}

	poin, err := r.PoinAktif(ctx, f)
	if err != nil {
		return out, fmt.Errorf("laporan poin aktif: %w", err)
	}
	out.PoinAktif = poin.TotalPoin
	out.NilaiPoin = poin.NilaiRupiah
	out.MemberBerpoin = poin.MemberBerpoin
	out.PoinPerMember = poin.PerMember
	out.PoinPerCabang = poin.PerCabang

	tier, err := r.DistribusiTier(ctx)
	if err != nil {
		return out, fmt.Errorf("distribusi tingkat member: %w", err)
	}
	out.Tier = tier
	for _, t := range tier {
		out.TotalMember += t.Jumlah
		switch t.Tingkat {
		case "GOLD":
			out.MemberGold = t.Jumlah
		case "SILVER":
			out.MemberSilver = t.Jumlah
		}
	}

	pertumbuhan, err := r.PertumbuhanMember(ctx, f.Periode)
	if err != nil {
		return out, fmt.Errorf("pertumbuhan member: %w", err)
	}
	out.Pertumbuhan = pertumbuhan
	for _, p := range pertumbuhan {
		out.TotalMemberBaru += p.JumlahBaru
	}

	return out, nil
}

// BiayaPromo menghitung realisasi biaya setiap promo dari ledger transaksi_promo.
// Semua promo tetap tampil, termasuk yang belum pernah dipakai.
func (r *AdminRepo) BiayaPromo(ctx context.Context, f model.FilterLaporan) ([]model.LaporanBiayaPromo, error) {
	mulai, selesai := batasPeriode(f.Periode)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.kode, p.nama, p.jenis,
		       COALESCE(x.jumlah, 0) AS jumlah_pemakaian,
		       COALESCE(x.total_biaya, 0) AS total_biaya
		FROM promo p
		LEFT JOIN (
			SELECT tp.promo_id,
			       COUNT(*) AS jumlah,
			       CAST(COALESCE(SUM(tp.diskon), 0) AS SIGNED) AS total_biaya
			FROM transaksi_promo tp
			JOIN transaksi t ON t.id = tp.transaksi_id
			WHERE t.status = 'SELESAI'
			  AND (? = 0 OR t.cabang_id = ?)
			  AND t.created_at >= ? AND t.created_at < ?
			GROUP BY tp.promo_id
		) x ON x.promo_id = p.id
		ORDER BY total_biaya DESC, p.kode
	`, f.CabangID, f.CabangID, mulai, selesai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LaporanBiayaPromo
	for rows.Next() {
		var x model.LaporanBiayaPromo
		if err := rows.Scan(&x.Kode, &x.Nama, &x.Jenis, &x.JumlahPemakaian, &x.TotalBiaya); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// LaporanPoin adalah hasil agregat laporan poin aktif beserta rinciannya.
type LaporanPoin struct {
	TotalPoin     int
	NilaiRupiah   int
	MemberBerpoin int
	PerMember     []model.LaporanPoinMember
	PerCabang     []model.LaporanPoinCabang
}

// PoinAktif menghitung saldo poin yang masih berlaku beserta rincian per member
// dan per cabang asal perolehan poin. Batch poina tanpa transaksi (poin warisan)
// tidak punya atribut cabang sehingga hanya ikut dihitung saat seluruh cabang
// dan seluruh periode dipilih.
func (r *AdminRepo) PoinAktif(ctx context.Context, f model.FilterLaporan) (LaporanPoin, error) {
	var out LaporanPoin
	mulai, selesai := batasPeriode(f.Periode)

	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT member_id),
		       CAST(COALESCE(SUM(sisa), 0) AS SIGNED),
		       CAST(COALESCE(SUM(sisa) * ?, 0) AS SIGNED)
		FROM poin_member p
		WHERE p.status = 'AKTIF' AND p.kadaluarsa_at > NOW() AND p.sisa > 0
		  AND (? = 0 OR (p.transaksi_id IS NOT NULL AND EXISTS (
		        SELECT 1 FROM transaksi t WHERE t.id = p.transaksi_id AND t.cabang_id = ?)))
		  AND p.diperoleh_at >= ? AND p.diperoleh_at < ?
	`, nilaiPoinPerPoin, f.CabangID, f.CabangID, mulai, selesai).
		Scan(&out.MemberBerpoin, &out.TotalPoin, &out.NilaiRupiah)
	if err != nil {
		return out, err
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT m.nomor_member, m.nama, m.tingkat,
		       COUNT(p.id) AS jumlah_batch,
		       CAST(COALESCE(SUM(p.sisa), 0) AS SIGNED),
		       CAST(COALESCE(SUM(p.sisa) * ?, 0) AS SIGNED),
		       DATE_FORMAT(MIN(p.kadaluarsa_at), '%d %b %Y')
		FROM poin_member p
		JOIN member m ON m.id = p.member_id
		WHERE p.status = 'AKTIF' AND p.kadaluarsa_at > NOW() AND p.sisa > 0
		  AND (? = 0 OR (p.transaksi_id IS NOT NULL AND EXISTS (
		        SELECT 1 FROM transaksi t WHERE t.id = p.transaksi_id AND t.cabang_id = ?)))
		  AND p.diperoleh_at >= ? AND p.diperoleh_at < ?
		GROUP BY m.id, m.nomor_member, m.nama, m.tingkat
		ORDER BY SUM(p.sisa) DESC
	`, nilaiPoinPerPoin, f.CabangID, f.CabangID, mulai, selesai)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	for rows.Next() {
		var x model.LaporanPoinMember
		if err := rows.Scan(&x.NomorMember, &x.Nama, &x.Tingkat, &x.JumlahBatch, &x.TotalPoin, &x.NilaiRupiah, &x.KedaluarsaTerdekat); err != nil {
			return out, err
		}
		out.PerMember = append(out.PerMember, x)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	rows.Close()

	rows2, err := r.DB.QueryContext(ctx, `
		SELECT c.nama,
		       CAST(COALESCE(SUM(p.sisa), 0) AS SIGNED),
		       CAST(COALESCE(SUM(p.sisa) * ?, 0) AS SIGNED)
		FROM poin_member p
		JOIN transaksi t ON t.id = p.transaksi_id
		JOIN cabang c ON c.id = t.cabang_id
		WHERE p.status = 'AKTIF' AND p.kadaluarsa_at > NOW() AND p.sisa > 0
		  AND (? = 0 OR t.cabang_id = ?)
		  AND p.diperoleh_at >= ? AND p.diperoleh_at < ?
		GROUP BY c.id, c.nama
		UNION ALL
		SELECT 'Tanpa Atribusi Cabang',
		       CAST(COALESCE(SUM(p.sisa), 0) AS SIGNED),
		       CAST(COALESCE(SUM(p.sisa) * ?, 0) AS SIGNED)
		FROM poin_member p
		WHERE p.transaksi_id IS NULL
		  AND p.status = 'AKTIF' AND p.kadaluarsa_at > NOW() AND p.sisa > 0
		  AND ? = 0
		  AND p.diperoleh_at >= ? AND p.diperoleh_at < ?
	`, nilaiPoinPerPoin, f.CabangID, f.CabangID, mulai, selesai,
		nilaiPoinPerPoin, f.CabangID, mulai, selesai)
	if err != nil {
		return out, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var x model.LaporanPoinCabang
		if err := rows2.Scan(&x.NamaCabang, &x.TotalPoin, &x.NilaiRupiah); err != nil {
			return out, err
		}
		out.PerCabang = append(out.PerCabang, x)
	}
	return out, rows2.Err()
}

// DistribusiTier menghitung sebaran member per tingkat keanggotaan.
func (r *AdminRepo) DistribusiTier(ctx context.Context) ([]model.LaporanTier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tingkat, COUNT(*) FROM member WHERE aktif = 1 GROUP BY tingkat
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LaporanTier
	total := 0
	for rows.Next() {
		var x model.LaporanTier
		if err := rows.Scan(&x.Tingkat, &x.Jumlah); err != nil {
			return nil, err
		}
		total += x.Jumlah
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Urutan tingkat yang konsisten: GOLD, SILVER, lalu lainnya.
	sort.SliceStable(out, func(i, j int) bool {
		return urutanTingkat(out[i].Tingkat) < urutanTingkat(out[j].Tingkat)
	})
	for i := range out {
		if total > 0 {
			out[i].Persen = float64(out[i].Jumlah) / float64(total) * 100
		}
	}
	return out, nil
}

// urutanTingkat memberi bobot pengurutan tingkat keanggotaan.
func urutanTingkat(tingkat string) int {
	switch tingkat {
	case "GOLD":
		return 0
	case "SILVER":
		return 1
	default:
		return 2
	}
}

// PertumbuhanMember menghitung jumlah member baru per bulan pada jendela 12 bulan
// terakhir, atau 12 bulan yang berakhir pada bulan yang difilter. Sumbu bulan
// dibangun di Go supaya bulan tanpa member baru tetap terlihat pada grafik.
func (r *AdminRepo) PertumbuhanMember(ctx context.Context, periode string) ([]model.LaporanPertumbuhan, error) {
	akhir, err := r.bulanServer(ctx)
	if err != nil {
		return nil, err
	}
	if periode != "" {
		akhir = periode
	}

	bulan := make([]string, 0, jumlahBulanGrafik)
	for i := jumlahBulanGrafik - 1; i >= 0; i-- {
		bulan = append(bulan, geserBulan(akhir, -i))
	}
	mulai, selesai := batasBulan(bulan[0], akhir)

	rows, err := r.DB.QueryContext(ctx, `
		SELECT DATE_FORMAT(created_at, '%Y-%m') AS bulan, COUNT(*)
		FROM member
		WHERE created_at >= ? AND created_at < ?
		GROUP BY bulan
	`, mulai, selesai)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jumlah := make(map[string]int)
	for rows.Next() {
		var b string
		var n int
		if err := rows.Scan(&b, &n); err != nil {
			return nil, err
		}
		jumlah[b] = n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tertinggi := 0
	for _, n := range jumlah {
		if n > tertinggi {
			tertinggi = n
		}
	}

	out := make([]model.LaporanPertumbuhan, 0, len(bulan))
	kumulatif := 0
	for _, b := range bulan {
		n := jumlah[b]
		kumulatif += n
		persen := 0.0
		if tertinggi > 0 {
			persen = float64(n) / float64(tertinggi) * 100
		}
		out = append(out, model.LaporanPertumbuhan{
			Bulan:          b,
			LabelBulan:     labelBulan(b),
			JumlahBaru:     n,
			TotalKumulatif: kumulatif,
			Persen:         persen,
		})
	}
	return out, nil
}

// PeriodeTersedia mengembalikan daftar bulan (YYYY-MM) yang memiliki transaksi
// selesai, dipakai untuk mengisi pilihan filter periode.
func (r *AdminRepo) PeriodeTersedia(ctx context.Context) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT DISTINCT DATE_FORMAT(created_at, '%Y-%m')
		FROM transaksi WHERE status = 'SELESAI' ORDER BY 1 DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// bulanServer mengambil bulan berjalan menurut waktu server database.
func (r *AdminRepo) bulanServer(ctx context.Context) (string, error) {
	var s string
	err := r.DB.QueryRowContext(ctx, "SELECT DATE_FORMAT(CURDATE(), '%Y-%m')").Scan(&s)
	return s, err
}

// batasPeriode mengubah filter bulan (YYYY-MM) menjadi rentang waktu inklusif-exclusive.
// Periode kosong berarti tanpa batas.
func batasPeriode(periode string) (string, string) {
	if periode == "" {
		return batasAwal, batasAkhir
	}
	t, err := time.Parse("2006-01", periode)
	if err != nil {
		return batasAwal, batasAkhir
	}
	return t.Format("2006-01") + "-01 00:00:00", t.AddDate(0, 1, 0).Format("2006-01") + "-01 00:00:00"
}

// batasBulan mengubah rentang bulan inklusif menjadi rentang waktu inklusif-exclusive.
func batasBulan(dari, sampai string) (string, string) {
	return dari + "-01 00:00:00", geserBulan(sampai, 1) + "-01 00:00:00"
}

// geserBulan menggeser format bulan YYYY-MM sebanyak n bulan.
func geserBulan(bulan string, n int) string {
	t, err := time.Parse("2006-01", bulan)
	if err != nil {
		return bulan
	}
	return t.AddDate(0, n, 0).Format("2006-01")
}

// labelBulan mengubah YYYY-MM menjadi label yang enak dibaca, misal "Nov 2026".
func labelBulan(bulan string) string {
	t, err := time.Parse("2006-01", bulan)
	if err != nil {
		return bulan
	}
	return t.Format("Jan 2006")
}
