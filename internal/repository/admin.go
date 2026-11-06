package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

type AdminRepo struct{ DB *sql.DB }

func (r *AdminRepo) Cabang(ctx context.Context) ([]model.Cabang, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,kode,nama,alamat,COALESCE(telepon,''),aktif FROM cabang ORDER BY id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Cabang
	for rows.Next() {
		var x model.Cabang
		if e = rows.Scan(&x.ID, &x.Kode, &x.Nama, &x.Alamat, &x.Telepon, &x.Aktif); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *AdminRepo) TambahCabang(ctx context.Context, x model.Cabang) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO cabang(kode,nama,alamat,telepon) VALUES(?,?,?,?)", x.Kode, x.Nama, x.Alamat, x.Telepon)
	return e
}

func (r *AdminRepo) Member(ctx context.Context) ([]model.Member, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,nomor_member,nama,no_hp,COALESCE(email,''),COALESCE(alamat,''),tingkat,total_belanja_3_bulan,aktif,created_at FROM member ORDER BY id DESC")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Member
	for rows.Next() {
		var x model.Member
		if e = rows.Scan(&x.ID, &x.Nomor, &x.Nama, &x.NoHP, &x.Email, &x.Alamat, &x.Tingkat, &x.Total, &x.Aktif, &x.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *AdminRepo) TambahMember(ctx context.Context, x model.Member) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO member(nomor_member,nama,no_hp,email,alamat) VALUES(?,?,?,?,?)", x.Nomor, x.Nama, x.NoHP, x.Email, x.Alamat)
	return e
}

func (r *AdminRepo) Promo(ctx context.Context) ([]model.Promo, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,kode,nama,jenis,nilai,minimal_belanja,COALESCE(maksimal_pemakaian,0),terpakai,mulai,berakhir,aktif FROM promo ORDER BY id DESC")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Promo
	for rows.Next() {
		var x model.Promo
		if e = rows.Scan(&x.ID, &x.Kode, &x.Nama, &x.Jenis, &x.Nilai, &x.Minimal, &x.Maksimal, &x.Terpakai, &x.Mulai, &x.Berakhir, &x.Aktif); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *AdminRepo) TambahPromo(ctx context.Context, x model.Promo) error {
	tx, e := r.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	res, e := tx.ExecContext(ctx, "INSERT INTO promo(kode,nama,jenis,nilai,minimal_belanja,maksimal_pemakaian,mulai,berakhir,mulai_jam,berakhir_jam) VALUES(?,?,?,?,?,?,?,?,?,?)", x.Kode, x.Nama, x.Jenis, x.Nilai, x.Minimal, x.Maksimal, x.Mulai, x.Berakhir, nullString(x.MulaiJam), nullString(x.BerakhirJam))
	if e != nil {
		tx.Rollback()
		return e
	}
	id, e := res.LastInsertId()
	if e != nil {
		tx.Rollback()
		return e
	}
	if x.CabangID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_cabang(promo_id,cabang_id) VALUES(?,?)", id, x.CabangID); e != nil {
			tx.Rollback()
			return e
		}
	}
	if x.TargetProdukID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_produk(promo_id,produk_id) VALUES(?,?)", id, x.TargetProdukID); e != nil {
			tx.Rollback()
			return e
		}
	}
	if x.TargetKategoriID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_kategori(promo_id,kategori_id) VALUES(?,?)", id, x.TargetKategoriID); e != nil {
			tx.Rollback()
			return e
		}
	}
	return tx.Commit()
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *AdminRepo) Voucher(ctx context.Context) ([]model.Voucher, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT v.id,v.promo_id,v.kode,v.kuota,v.terpakai,p.nama FROM voucher v JOIN promo p ON p.id=v.promo_id ORDER BY v.id DESC")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Voucher
	for rows.Next() {
		var x model.Voucher
		if e = rows.Scan(&x.ID, &x.PromoID, &x.Kode, &x.Kuota, &x.Terpakai, &x.PromoNama); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *AdminRepo) TambahVoucher(ctx context.Context, x model.Voucher) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO voucher(promo_id,kode,kuota) VALUES(?,?,?)", x.PromoID, x.Kode, x.Kuota)
	return e
}

func (r *AdminRepo) Ringkasan(ctx context.Context) (model.RingkasanData, error) {
	var out model.RingkasanData
	_ = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM produk WHERE deleted_at IS NULL").Scan(&out.TotalProduk)
	_ = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM cabang WHERE aktif = 1").Scan(&out.TotalCabang)
	_ = r.DB.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(tingkat='GOLD'),0), COALESCE(SUM(tingkat='SILVER'),0) FROM member").Scan(&out.TotalMember, &out.MemberGold, &out.MemberSilver)
	_ = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM promo WHERE aktif = 1").Scan(&out.TotalPromo)
	_ = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM voucher").Scan(&out.TotalVoucher)
	return out, nil
}

// MemberByEmail mencari data member berdasarkan email pengguna login
func (r *AdminRepo) MemberByEmail(ctx context.Context, email string) (model.Member, error) {
	var x model.Member
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, nomor_member, nama, no_hp, COALESCE(email,''), COALESCE(alamat,''), tingkat, total_belanja_3_bulan, aktif, created_at
		FROM member
		WHERE email = ? OR no_hp = ?
		LIMIT 1
	`, email, email).Scan(&x.ID, &x.Nomor, &x.Nama, &x.NoHP, &x.Email, &x.Alamat, &x.Tingkat, &x.Total, &x.Aktif, &x.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		// Fallback ke member pertama jika belum ada binding spesifik
		err = r.DB.QueryRowContext(ctx, `
			SELECT id, nomor_member, nama, no_hp, COALESCE(email,''), COALESCE(alamat,''), tingkat, total_belanja_3_bulan, aktif, created_at
			FROM member
			ORDER BY id ASC LIMIT 1
		`).Scan(&x.ID, &x.Nomor, &x.Nama, &x.NoHP, &x.Email, &x.Alamat, &x.Tingkat, &x.Total, &x.Aktif, &x.CreatedAt)
	}

	return x, err
}

// PoinByMemberID mengambil rincian saldo poin aktif dan batch FEFO
func (r *AdminRepo) PoinByMemberID(ctx context.Context, memberID int) (int, []model.PoinDetail, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, member_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status
		FROM poin_member
		WHERE member_id = ?
		ORDER BY kadaluarsa_at ASC
	`, memberID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var totalAktif int
	var list []model.PoinDetail
	for rows.Next() {
		var p model.PoinDetail
		if err := rows.Scan(&p.ID, &p.MemberID, &p.Jumlah, &p.Sisa, &p.DiperolehAt, &p.KadaluarsaAt, &p.Status); err != nil {
			return 0, nil, err
		}
		if p.Status == "AKTIF" && p.Sisa > 0 {
			totalAktif += p.Sisa
		}
		list = append(list, p)
	}
	return totalAktif, list, rows.Err()
}

// SemuaTransaksi mengambil riwayat transaksi untuk kasir/admin
func (r *AdminRepo) SemuaTransaksi(ctx context.Context) ([]model.Transaksi, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT t.id, t.cabang_id, COALESCE(c.nama, 'Cabang'), COALESCE(t.member_id, 0),
		       COALESCE(m.nama, 'Pelanggan Umum'), COALESCE(m.nomor_member, '-'),
		       t.subtotal, t.total_diskon, t.total_bayar, t.status, t.created_at
		FROM transaksi t
		LEFT JOIN cabang c ON c.id = t.cabang_id
		LEFT JOIN member m ON m.id = t.member_id
		ORDER BY t.id DESC LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Transaksi
	for rows.Next() {
		var x model.Transaksi
		if err := rows.Scan(&x.ID, &x.CabangID, &x.CabangNama, &x.MemberID, &x.MemberNama, &x.MemberNomor, &x.Subtotal, &x.TotalDiskon, &x.TotalBayar, &x.Status, &x.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

// TransaksiByMemberID mengambil riwayat transaksi khusus member tertentu
func (r *AdminRepo) TransaksiByMemberID(ctx context.Context, memberID int) ([]model.Transaksi, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT t.id, t.cabang_id, COALESCE(c.nama, 'Cabang'), t.member_id,
		       COALESCE(m.nama, ''), COALESCE(m.nomor_member, ''),
		       t.subtotal, t.total_diskon, t.total_bayar, t.status, t.created_at
		FROM transaksi t
		LEFT JOIN cabang c ON c.id = t.cabang_id
		LEFT JOIN member m ON m.id = t.member_id
		WHERE t.member_id = ?
		ORDER BY t.id DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Transaksi
	for rows.Next() {
		var x model.Transaksi
		if err := rows.Scan(&x.ID, &x.CabangID, &x.CabangNama, &x.MemberID, &x.MemberNama, &x.MemberNomor, &x.Subtotal, &x.TotalDiskon, &x.TotalBayar, &x.Status, &x.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, x)
	}
	return list, rows.Err()
}

// Laporan menghasilkan 3 rekapitulasi sesuai soal Topik 5
func (r *AdminRepo) Laporan(ctx context.Context) (model.LaporanData, error) {
	var out model.LaporanData

	// 1. Biaya setiap promo
	rows, err := r.DB.QueryContext(ctx, `
		SELECT kode, nama, jenis, terpakai, nilai, (terpakai * nilai) AS total_biaya
		FROM promo
		ORDER BY terpakai DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var lp model.LaporanPromo
			if err := rows.Scan(&lp.Kode, &lp.Nama, &lp.Jenis, &lp.Terpakai, &lp.Nilai, &lp.TotalBiaya); err == nil {
				out.ListPromo = append(out.ListPromo, lp)
			}
		}
	}

	// 2. Poin aktif & nilai rupiah (1 poin = Rp1.000)
	_ = r.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(sisa), 0), COALESCE(SUM(sisa) * 1000, 0)
		FROM poin_member
		WHERE status = 'AKTIF' AND kadaluarsa_at > NOW()
	`).Scan(&out.TotalPoinAktif, &out.NilaiRupiah)

	// 3. Perkembangan member
	_ = r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(tingkat = 'GOLD'), 0), COALESCE(SUM(tingkat = 'SILVER'), 0)
		FROM member
	`).Scan(&out.TotalMember, &out.MemberGold, &out.MemberSilver)

	return out, nil
}

// TambahTransaksiKasir memproses transaksi kasir POS
func (r *AdminRepo) TambahTransaksiKasir(ctx context.Context, cabangID, memberID, subtotal, diskon, bayar, poinTukar, poinDapat int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var mID any = nil
	if memberID > 0 {
		mID = memberID
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO transaksi (cabang_id, member_id, subtotal, total_diskon, total_bayar, status)
		VALUES (?, ?, ?, ?, ?, 'SELESAI')
	`, cabangID, mID, subtotal, diskon, bayar)
	if err != nil {
		tx.Rollback()
		return err
	}

	transaksiID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Jika ada penukaran poin, kurangi saldo poin terdekat (FEFO)
	if memberID > 0 && poinTukar > 0 {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, sisa
			FROM poin_member
			WHERE member_id = ? AND status = 'AKTIF' AND sisa > 0 AND kadaluarsa_at > NOW()
			ORDER BY kadaluarsa_at ASC
		`, memberID)
		if err == nil {
			sisaTukar := poinTukar
			var updates [][2]int
			for rows.Next() {
				var pID, sisa int
				if rows.Scan(&pID, &sisa) == nil {
					if sisaTukar <= 0 {
						break
					}
					potong := sisa
					if potong > sisaTukar {
						potong = sisaTukar
					}
					updates = append(updates, [2]int{pID, sisa - potong})
					sisaTukar -= potong
				}
			}
			rows.Close()
			for _, u := range updates {
				_, _ = tx.ExecContext(ctx, "UPDATE poin_member SET sisa = ? WHERE id = ?", u[1], u[0])
			}
		}
	}

	// Jika ada perolehan poin baru, masukkan ke poin_member dengan masa berlaku 6 bulan
	if memberID > 0 && poinDapat > 0 {
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status)
			VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')
		`, memberID, transaksiID, poinDapat, poinDapat)
	}

	return tx.Commit()
}

// BatalkanTransaksi membatalkan pesanan dan mengembalikan poin
func (r *AdminRepo) BatalkanTransaksi(ctx context.Context, transaksiID int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 1. Ubah status transaksi
	_, err = tx.ExecContext(ctx, "UPDATE transaksi SET status = 'DIBATALKAN' WHERE id = ?", transaksiID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Tarik kembali poin yang diperoleh dari transaksi ini
	_, _ = tx.ExecContext(ctx, "UPDATE poin_member SET status = 'DIBATALKAN', sisa = 0 WHERE transaksi_id = ?", transaksiID)

	return tx.Commit()
}
