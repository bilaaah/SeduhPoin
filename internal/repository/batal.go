package repository

import (
	"context"
	"database/sql"
	"errors"
)

func (r *AdminRepo) BatalkanTransaksi(ctx context.Context, transaksiID int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Pastikan transaksi tersedia dan masih SELESAI.
	var status string
	var voucherID sql.NullInt64

	err = tx.QueryRowContext(ctx, `
		SELECT status, voucher_id
		FROM transaksi
		WHERE id = ?
		FOR UPDATE
	`, transaksiID).Scan(&status, &voucherID)

	if err != nil {
		return err
	}

	if status != "SELESAI" {
		return errors.New("transaksi tidak dapat dibatalkan")
	}

	// 2. Kembalikan stok produk.
	_, err = tx.ExecContext(ctx, `
		UPDATE produk p
		JOIN transaksi_item ti ON ti.produk_id = p.id
		SET p.stok = p.stok + ti.jumlah
		WHERE ti.transaksi_id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	// 3. Kembalikan poin yang pernah ditukarkan.
	rows, err := tx.QueryContext(ctx, `
		SELECT poin_member_id, jumlah_dipakai
		FROM transaksi_poin_ditukar
		WHERE transaksi_id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	type pengembalianPoin struct {
		ID     int
		Jumlah int
	}

	var daftarPoin []pengembalianPoin

	for rows.Next() {
		var p pengembalianPoin
		if err := rows.Scan(&p.ID, &p.Jumlah); err != nil {
			rows.Close()
			return err
		}
		daftarPoin = append(daftarPoin, p)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}

	for _, p := range daftarPoin {
		_, err = tx.ExecContext(ctx, `
			UPDATE poin_member
			SET sisa = sisa + ?
			WHERE id = ?
		`, p.Jumlah, p.ID)
		if err != nil {
			return err
		}
	}

	// 4. Batalkan poin reward dari transaksi ini.
	_, err = tx.ExecContext(ctx, `
		UPDATE poin_member
		SET status = 'DIBATALKAN', sisa = 0
		WHERE transaksi_id = ?
		  AND status = 'AKTIF'
	`, transaksiID)
	if err != nil {
		return err
	}

	// 5. Kembalikan kuota voucher jika digunakan.
	// 5. Kembalikan kuota voucher dan pemakaian promo jika digunakan.
if voucherID.Valid {
    _, err = tx.ExecContext(ctx, `
        UPDATE voucher
        SET terpakai = GREATEST(terpakai - 1, 0)
        WHERE id = ?
    `, voucherID.Int64)
    if err != nil {
        return err
    }

    _, err = tx.ExecContext(ctx, `
        UPDATE promo p
        JOIN voucher v ON v.promo_id = p.id
        SET p.terpakai = GREATEST(p.terpakai - 1, 0)
        WHERE v.id = ?
    `, voucherID.Int64)
    if err != nil {
        return err
    }
}

	// 6. Ubah status transaksi.
	_, err = tx.ExecContext(ctx, `
		UPDATE transaksi
		SET status = 'DIBATALKAN'
		WHERE id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
