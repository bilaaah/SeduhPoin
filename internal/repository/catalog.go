package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

func (r *AdminRepo) Kategori(ctx context.Context) ([]model.Kategori, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id,nama FROM kategori ORDER BY nama")
	if err != nil {
		return nil, fmt.Errorf("daftar kategori: %w", err)
	}
	defer rows.Close()
	var hasil []model.Kategori
	for rows.Next() {
		var x model.Kategori
		if err := rows.Scan(&x.ID, &x.Nama); err != nil {
			return nil, err
		}
		hasil = append(hasil, x)
	}
	return hasil, rows.Err()
}

func (r *AdminRepo) Produk(ctx context.Context) ([]model.Produk, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT p.id,COALESCE(p.kategori_id,0),COALESCE(k.nama,''),p.nama,p.harga,p.stok,COALESCE(p.keterangan,''),p.aktif FROM produk p LEFT JOIN kategori k ON k.id=p.kategori_id WHERE p.deleted_at IS NULL ORDER BY p.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("daftar produk: %w", err)
	}
	defer rows.Close()
	var hasil []model.Produk
	for rows.Next() {
		var x model.Produk
		if err := rows.Scan(&x.ID, &x.KategoriID, &x.Kategori, &x.Nama, &x.Harga, &x.Stok, &x.Keterangan, &x.Aktif); err != nil {
			return nil, err
		}
		hasil = append(hasil, x)
	}
	return hasil, rows.Err()
}

func (r *AdminRepo) TambahProduk(ctx context.Context, x model.Produk) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO produk(kategori_id,nama,harga,stok,keterangan) VALUES(?,?,?,?,?)", x.KategoriID, x.Nama, x.Harga, x.Stok, x.Keterangan)
	return err
}

func (r *AdminRepo) UbahProduk(ctx context.Context, x model.Produk) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE produk SET kategori_id=?,nama=?,harga=?,stok=?,keterangan=?,updated_at=NOW() WHERE id=? AND deleted_at IS NULL", x.KategoriID, x.Nama, x.Harga, x.Stok, x.Keterangan, x.ID)
	if err != nil {
		return err
	}
	return pastikanAdaBaris(res)
}

func (r *AdminRepo) HapusProduk(ctx context.Context, id int) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE produk SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	return pastikanAdaBaris(res)
}

func pastikanAdaBaris(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
