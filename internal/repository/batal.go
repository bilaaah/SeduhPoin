package repository

import (
	"context"
	"database/sql"
	"errors"
)

// BatalkanTransaksi membatalkan pesanan kasir secara atomik (Lingkup 4 - Hilal):
// 1. Validasi transaksi berstatus SELESAI
// 2. Mengembalikan stok produk berdasarkan transaksi_item (qty)
// 3. Membatalkan poin reward yang diperoleh transaksi ini (void)
// 4. Mengembalikan poin yang pernah ditukarkan ke saldo aktif member (masa berlaku 6 bulan)
// 5. Mengembalikan kuota voucher dan counter promo yang digunakan
// 6. Mengubah status transaksi menjadi DIBATALKAN
// 7. Menghitung ulang total belanja 3 bulan & tingkat member
func (r *AdminRepo) BatalkanTransaksi(ctx context.Context, transaksiID int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Pastikan transaksi ada, kunci baris, dan cek status SELESAI
	var status string
	var memberID sql.NullInt64
	var voucherID sql.NullInt64
	var poinDigunakan int

	err = tx.QueryRowContext(ctx, `
		SELECT status, member_id, voucher_id, poin_digunakan
		FROM transaksi
		WHERE id = ?
		FOR UPDATE
	`, transaksiID).Scan(&status, &memberID, &voucherID, &poinDigunakan)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("transaksi tidak ditemukan")
	}
	if err != nil {
		return err
	}
	if status != "SELESAI" {
		return errors.New("transaksi tidak dapat dibatalkan atau sudah berstatus batal")
	}

	// 2. Kembalikan stok produk berdasarkan transaksi_item (ti.qty)
	_, err = tx.ExecContext(ctx, `
		UPDATE produk p
		JOIN transaksi_item ti ON ti.produk_id = p.id
		SET p.stok = p.stok + ti.qty
		WHERE ti.transaksi_id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	// 3. Batalkan poin reward yang diperoleh transaksi ini (void)
	_, err = tx.ExecContext(ctx, `
		UPDATE poin_member
		SET status = 'DIBATALKAN', sisa = 0
		WHERE transaksi_id = ? AND status = 'AKTIF'
	`, transaksiID)
	if err != nil {
		return err
	}

	// 4. Kembalikan poin yang pernah ditukarkan ke saldo aktif member (masa berlaku 6 bulan)
	if memberID.Valid && memberID.Int64 > 0 && poinDigunakan > 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status)
			VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')
		`, memberID.Int64, transaksiID, poinDigunakan, poinDigunakan)
		if err != nil {
			return err
		}
	}

	// 5. Kembalikan kuota voucher jika transaksi menggunakan voucher
	if voucherID.Valid && voucherID.Int64 > 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE voucher SET terpakai = GREATEST(terpakai - 1, 0) WHERE id = ?
		`, voucherID.Int64)
		if err != nil {
			return err
		}
	}

	// Kembalikan pemakaian promo dari ledger transaksi_promo (Lingkup 6 Luhur)
	_, err = tx.ExecContext(ctx, `
		UPDATE promo p
		JOIN transaksi_promo tp ON tp.promo_id = p.id
		SET p.terpakai = GREATEST(p.terpakai - 1, 0)
		WHERE tp.transaksi_id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	// 6. Ubah status transaksi menjadi DIBATALKAN
	_, err = tx.ExecContext(ctx, `
		UPDATE transaksi SET status = 'DIBATALKAN' WHERE id = ?
	`, transaksiID)
	if err != nil {
		return err
	}

	// 7. Hitung ulang akumulasi belanja 3 bulan & tingkat member jika transaksi punya member
	if memberID.Valid && memberID.Int64 > 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE member m
			LEFT JOIN (
				SELECT member_id, SUM(total_bayar) AS total
				FROM transaksi
				WHERE status = 'SELESAI' AND member_id = ?
				  AND created_at >= NOW() - INTERVAL 3 MONTH
				GROUP BY member_id
			) t ON t.member_id = m.id
			SET m.total_belanja_3_bulan = COALESCE(t.total, 0),
			    m.tingkat = IF(COALESCE(t.total, 0) >= 1500000, 'GOLD', 'SILVER')
			WHERE m.id = ?
		`, memberID.Int64, memberID.Int64)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
