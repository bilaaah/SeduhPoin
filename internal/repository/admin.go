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

func (r *AdminRepo) UbahCabang(ctx context.Context, x model.Cabang) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE cabang SET nama = ?, alamat = ?, telepon = ?, aktif = ? WHERE id = ?", x.Nama, x.Alamat, x.Telepon, x.Aktif, x.ID)
	return e
}

func (r *AdminRepo) ToggleCabang(ctx context.Context, id int) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE cabang SET aktif = NOT aktif WHERE id = ?", id)
	return e
}

func (r *AdminRepo) Member(ctx context.Context) ([]model.Member, error) {
	rows, e := r.DB.QueryContext(ctx, `
		SELECT m.id, m.nomor_member, m.nama, m.no_hp, COALESCE(m.email,''), COALESCE(m.alamat,''),
		       m.tingkat, m.total_belanja_3_bulan, m.aktif, m.created_at,
		       COALESCE((SELECT SUM(pm.sisa) FROM poin_member pm WHERE pm.member_id = m.id AND pm.status = 'AKTIF' AND pm.sisa > 0 AND pm.kadaluarsa_at > NOW()), 0)
		FROM member m ORDER BY m.id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Member
	for rows.Next() {
		var x model.Member
		if e = rows.Scan(&x.ID, &x.Nomor, &x.Nama, &x.NoHP, &x.Email, &x.Alamat, &x.Tingkat, &x.Total, &x.Aktif, &x.CreatedAt, &x.SaldoPoin); e != nil {
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

func (r *AdminRepo) UbahMember(ctx context.Context, x model.Member) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE member SET nama = ?, no_hp = ?, email = ?, alamat = ?, aktif = ? WHERE id = ?", x.Nama, x.NoHP, x.Email, x.Alamat, x.Aktif, x.ID)
	return e
}

func (r *AdminRepo) ToggleMember(ctx context.Context, id int) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE member SET aktif = NOT aktif WHERE id = ?", id)
	return e
}

func (r *AdminRepo) Promo(ctx context.Context) ([]model.Promo, error) {
	rows, e := r.DB.QueryContext(ctx, `
		SELECT id, kode, nama, jenis, nilai, minimal_belanja, COALESCE(maksimal_pemakaian,0), terpakai,
		       mulai, berakhir, COALESCE(DATE_FORMAT(mulai_jam,'%H:%i'),''), COALESCE(DATE_FORMAT(berakhir_jam,'%H:%i'),''), aktif
		FROM promo ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Promo
	idx := make(map[int]int)
	for rows.Next() {
		var x model.Promo
		if e = rows.Scan(&x.ID, &x.Kode, &x.Nama, &x.Jenis, &x.Nilai, &x.Minimal, &x.Maksimal, &x.Terpakai, &x.Mulai, &x.Berakhir, &x.MulaiJam, &x.BerakhirJam, &x.Aktif); e != nil {
			return nil, e
		}
		idx[x.ID] = len(out)
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if r2, err := r.DB.QueryContext(ctx, "SELECT promo_id, produk_id FROM promo_produk"); err == nil {
		defer r2.Close()
		for r2.Next() {
			var pid, prid int
			if r2.Scan(&pid, &prid) == nil {
				if i, ok := idx[pid]; ok {
					out[i].ProdukIDs = append(out[i].ProdukIDs, prid)
				}
			}
		}
	}
	if r3, err := r.DB.QueryContext(ctx, "SELECT promo_id, kategori_id FROM promo_kategori"); err == nil {
		defer r3.Close()
		for r3.Next() {
			var pid, kid int
			if r3.Scan(&pid, &kid) == nil {
				if i, ok := idx[pid]; ok {
					out[i].KategoriIDs = append(out[i].KategoriIDs, kid)
				}
			}
		}
	}
	if r4, err := r.DB.QueryContext(ctx, "SELECT promo_id, cabang_id FROM promo_cabang"); err == nil {
		defer r4.Close()
		for r4.Next() {
			var pid, cid int
			if r4.Scan(&pid, &cid) == nil {
				if i, ok := idx[pid]; ok {
					out[i].CabangIDs = append(out[i].CabangIDs, cid)
				}
			}
		}
	}
	return out, nil
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

func (r *AdminRepo) UbahPromo(ctx context.Context, x model.Promo) error {
	tx, e := r.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `
		UPDATE promo 
		SET nama=?, jenis=?, nilai=?, minimal_belanja=?, maksimal_pemakaian=?, mulai=?, berakhir=?, mulai_jam=?, berakhir_jam=? 
		WHERE id=?
	`, x.Nama, x.Jenis, x.Nilai, x.Minimal, x.Maksimal, x.Mulai, x.Berakhir, nullString(x.MulaiJam), nullString(x.BerakhirJam), x.ID)
	if e != nil {
		tx.Rollback()
		return e
	}

	// Update relasi cabang jika ditentukan atau direset
	_, _ = tx.ExecContext(ctx, "DELETE FROM promo_cabang WHERE promo_id=?", x.ID)
	if x.CabangID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_cabang(promo_id,cabang_id) VALUES(?,?)", x.ID, x.CabangID); e != nil {
			tx.Rollback()
			return e
		}
	}

	// Update relasi produk
	_, _ = tx.ExecContext(ctx, "DELETE FROM promo_produk WHERE promo_id=?", x.ID)
	if x.TargetProdukID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_produk(promo_id,produk_id) VALUES(?,?)", x.ID, x.TargetProdukID); e != nil {
			tx.Rollback()
			return e
		}
	}

	// Update relasi kategori
	_, _ = tx.ExecContext(ctx, "DELETE FROM promo_kategori WHERE promo_id=?", x.ID)
	if x.TargetKategoriID > 0 {
		if _, e = tx.ExecContext(ctx, "INSERT INTO promo_kategori(promo_id,kategori_id) VALUES(?,?)", x.ID, x.TargetKategoriID); e != nil {
			tx.Rollback()
			return e
		}
	}

	return tx.Commit()
}

func (r *AdminRepo) TogglePromo(ctx context.Context, id int) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE promo SET aktif = NOT aktif WHERE id = ?", id)
	return e
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

func (r *AdminRepo) UbahVoucher(ctx context.Context, id, kuota int) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE voucher SET kuota = ? WHERE id = ?", kuota, id)
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
