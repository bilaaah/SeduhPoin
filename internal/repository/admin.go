package repository

import (
	"context"
	"database/sql"
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
