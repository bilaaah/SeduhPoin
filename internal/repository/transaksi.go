package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

// ErrValidasi menandai kesalahan input pengguna agar handler membalas 400,
// bukan 500.
type ErrValidasi struct{ Pesan string }

func (e ErrValidasi) Error() string { return e.Pesan }

// TambahTransaksiKasir menyimpan transaksi penjualan POS sekaligus mencatat
// promo yang dipakai ke ledger transaksi_promo. Ledger inilah sumber kebenaran
// biaya promo pada Laporan Analitik, jadi setiap promo yang dipakai wajib
// memiliki nominal diskon yang dicatat di sini.
func (r *AdminRepo) TambahTransaksiKasir(ctx context.Context, in model.TransaksiInput) error {
	if in.CabangID <= 0 {
		return ErrValidasi{Pesan: "Cabang transaksi belum dipilih"}
	}
	if in.Subtotal <= 0 {
		return ErrValidasi{Pesan: "Subtotal belanja harus lebih dari nol"}
	}

	alokasi, diskon, err := gabungAlokasi(in.Promo)
	if err != nil {
		return err
	}
	if diskon > in.Subtotal {
		return ErrValidasi{Pesan: "Total diskon promo melebihi subtotal belanja"}
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 1. Tukar poin member dengan sistem FEFO (batch terdekat kedaluwarsa dulu).
	poinTukar, err := tukarPoin(ctx, tx, in.MemberID, in.PoinDigunakan)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Simpan transaksi dengan diskon yang dihitung dari alokasi promo.
	bayar := in.Subtotal - diskon - poinTukar*nilaiPoinPerPoin
	if bayar < 0 {
		tx.Rollback()
		return ErrValidasi{Pesan: "Total poin yang ditukar terlalu besar untuk nilai transaksi"}
	}

	var mID, vID any
	if in.MemberID > 0 {
		mID = in.MemberID
	}
	if in.VoucherID > 0 {
		vID = in.VoucherID
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO transaksi (cabang_id, member_id, voucher_id, poin_digunakan, subtotal, total_diskon, total_bayar, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'SELESAI')
	`, in.CabangID, mID, vID, poinTukar, in.Subtotal, diskon, bayar)
	if err != nil {
		tx.Rollback()
		return err
	}
	transaksiID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Catat ledger promo, perbarui counter promo dan voucher.
	for _, a := range alokasi {
		var jenis string
		var aktif bool
		var maksimal, terpakai int
		err := tx.QueryRowContext(ctx, `
			SELECT jenis, aktif, COALESCE(maksimal_pemakaian, 0), terpakai FROM promo WHERE id = ?
		`, a.PromoID).Scan(&jenis, &aktif, &maksimal, &terpakai)
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Promo #%d tidak ditemukan", a.PromoID)}
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		if !aktif {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Promo #%d sedang tidak aktif", a.PromoID)}
		}
		if maksimal > 0 && terpakai >= maksimal {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Kuota promo #%d sudah habis", a.PromoID)}
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transaksi_promo (transaksi_id, promo_id, jenis, diskon) VALUES (?, ?, ?, ?)
		`, transaksiID, a.PromoID, jenis, a.Diskon); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE promo SET terpakai = terpakai + 1 WHERE id = ?", a.PromoID); err != nil {
			tx.Rollback()
			return err
		}

		if a.VoucherKode == "" {
			continue
		}
		var vIDvoucher int
		var vPromoID, kuota, vTerpakai int
		err = tx.QueryRowContext(ctx, `
			SELECT id, promo_id, kuota, terpakai FROM voucher WHERE kode = ?
		`, a.VoucherKode).Scan(&vIDvoucher, &vPromoID, &kuota, &vTerpakai)
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Kode voucher %s tidak ditemukan", a.VoucherKode)}
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		if vPromoID != a.PromoID {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Kode voucher %s tidak berlaku untuk promo yang dipilih", a.VoucherKode)}
		}
		if vTerpakai >= kuota {
			tx.Rollback()
			return ErrValidasi{Pesan: fmt.Sprintf("Kuota voucher %s sudah habis", a.VoucherKode)}
		}
		if _, err := tx.ExecContext(ctx, "UPDATE voucher SET terpakai = terpakai + 1 WHERE id = ?", vIDvoucher); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 4. Simpan perolehan poin baru dengan masa berlaku 6 bulan.
	if in.MemberID > 0 && in.PoinDiperoleh > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status)
			VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')
		`, in.MemberID, transaksiID, in.PoinDiperoleh, in.PoinDiperoleh); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// BatalkanTransaksi membatalkan pesanan dan menarik kembali poin yang diperoleh
// dari transaksi tersebut, sekaligus mengembalikan counter promo dan voucher.
func (r *AdminRepo) BatalkanTransaksi(ctx context.Context, transaksiID int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE transaksi SET status = 'DIBATALKAN' WHERE id = ?", transaksiID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE poin_member SET status = 'DIBATALKAN', sisa = 0 WHERE transaksi_id = ?
	`, transaksiID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE promo p JOIN transaksi_promo tp ON tp.promo_id = p.id
		SET p.terpakai = GREATEST(p.terpakai - 1, 0)
		WHERE tp.transaksi_id = ?
	`, transaksiID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE voucher v JOIN transaksi t ON t.voucher_id = v.id
		SET v.terpakai = GREATEST(v.terpakai - 1, 0)
		WHERE t.id = ?
	`, transaksiID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// gabungAlokasi menjumlahkan diskon per promo dan memvalidasi nominalnya.
func gabungAlokasi(promo []model.AlokasiPromo) ([]model.AlokasiPromo, int, error) {
	terpakai := make(map[int]int)
	voucher := make(map[int]string)
	var urut []int
	diskon := 0

	for _, a := range promo {
		if a.PromoID <= 0 {
			continue
		}
		if a.Diskon <= 0 {
			return nil, 0, ErrValidasi{Pesan: "Nominal diskon promo harus lebih dari nol"}
		}
		if _, ada := terpakai[a.PromoID]; !ada {
			urut = append(urut, a.PromoID)
		}
		terpakai[a.PromoID] += a.Diskon
		if a.VoucherKode != "" {
			voucher[a.PromoID] = a.VoucherKode
		}
		diskon += a.Diskon
	}

	var out []model.AlokasiPromo
	for _, id := range urut {
		out = append(out, model.AlokasiPromo{
			PromoID:     id,
			Diskon:      terpakai[id],
			VoucherKode: voucher[id],
		})
	}
	return out, diskon, nil
}

// tukarPoin mengonsumsi saldo poin member memakai FEFO dan mengembalikan jumlah
// poin yang benar-benar terpakai.
func tukarPoin(ctx context.Context, tx *sql.Tx, memberID, jumlah int) (int, error) {
	if memberID <= 0 || jumlah <= 0 {
		return 0, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, sisa
		FROM poin_member
		WHERE member_id = ? AND status = 'AKTIF' AND sisa > 0 AND kadaluarsa_at > NOW()
		ORDER BY kadaluarsa_at ASC
	`, memberID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	sisaTukar := jumlah
	typeUpdate := [][2]int{}
	for rows.Next() {
		var pID, sisa int
		if err := rows.Scan(&pID, &sisa); err != nil {
			return 0, err
		}
		if sisaTukar <= 0 {
			break
		}
		potong := sisa
		if potong > sisaTukar {
			potong = sisaTukar
		}
		typeUpdate = append(typeUpdate, [2]int{pID, sisa - potong})
		sisaTukar -= potong
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	rows.Close()

	for _, u := range typeUpdate {
		if _, err := tx.ExecContext(ctx, "UPDATE poin_member SET sisa = ? WHERE id = ?", u[1], u[0]); err != nil {
			return 0, err
		}
	}
	return jumlah - sisaTukar, nil
}
