package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/bilaaah/SeduhPoin/internal/kalkulasi"
	"github.com/bilaaah/SeduhPoin/internal/model"
)

// ErrValidasi menandai kesalahan input pengguna agar handler membalas 400,
// bukan 500.
type ErrValidasi struct{ Pesan string }

func (e ErrValidasi) Error() string { return e.Pesan }

// TambahTransaksiKasir menyimpan transaksi penjualan POS sekaligus mencatat
// item produk, pengurangan stok, promo kombinasi, penukaran poin FEFO,
// perolehan poin baru (6 bulan), serta ledger transaksi_promo.
func (r *AdminRepo) TambahTransaksiKasir(ctx context.Context, in model.TransaksiInput) error {
	if in.CabangID <= 0 {
		return ErrValidasi{Pesan: "Cabang transaksi belum dipilih"}
	}

	// Jika kasir memilih daftar produk (Lingkup 3 POS lengkap)
	if len(in.Items) > 0 {
		return r.prosesTransaksiPOS(ctx, in)
	}

	// Fallback jika masih memakai input subtotal manual
	if in.Subtotal <= 0 {
		return ErrValidasi{Pesan: "Pilih minimal 1 produk atau isi subtotal belanja"}
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

	poinTukar, err := tukarPoin(ctx, tx, in.MemberID, in.PoinDigunakan)
	if err != nil {
		tx.Rollback()
		return err
	}

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

// prosesTransaksiPOS menjalankan seluruh logika Lingkup 3 secara atomik:
// validasi stok produk, kalkulasi diskon promo kombinasi & aturan bentrok voucher,
// penukaran poin FEFO (maks 50%), pengurangan stok, penyimpanan transaksi_item,
// pencatatan ledger transaksi_promo, dan penerbitan poin baru (Gold 1.5x, 6 bulan).
func (r *AdminRepo) prosesTransaksiPOS(ctx context.Context, in model.TransaksiInput) error {
	// Gabungkan item dengan produk_id yang sama
	urutID := []int{}
	qtyPerProduk := make(map[int]int)
	for _, it := range in.Items {
		if it.ProdukID <= 0 || it.Qty <= 0 {
			continue
		}
		if _, ok := qtyPerProduk[it.ProdukID]; !ok {
			urutID = append(urutID, it.ProdukID)
		}
		qtyPerProduk[it.ProdukID] += it.Qty
	}
	if len(urutID) == 0 {
		return ErrValidasi{Pesan: "Pilih minimal 1 produk dengan jumlah lebih dari 0"}
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Kunci baris produk & validasi stok tersedia
	var items []kalkulasi.Item
	subtotal := 0
	for _, pid := range urutID {
		qty := qtyPerProduk[pid]
		var it kalkulasi.Item
		var stok int
		var aktif bool
		err := tx.QueryRowContext(ctx, `
			SELECT id, COALESCE(kategori_id, 0), nama, harga, stok, aktif
			FROM produk
			WHERE id = ? AND deleted_at IS NULL
			FOR UPDATE
		`, pid).Scan(&it.ProdukID, &it.KategoriID, &it.Nama, &it.Harga, &stok, &aktif)
		if errors.Is(err, sql.ErrNoRows) || !aktif {
			return ErrValidasi{Pesan: fmt.Sprintf("Produk #%d tidak ditemukan atau sedang nonaktif", pid)}
		}
		if err != nil {
			return err
		}
		if stok < qty {
			return ErrValidasi{Pesan: fmt.Sprintf("Stok %s tidak mencukupi (tersisa %d, diminta %d)", it.Nama, stok, qty)}
		}
		it.Qty = qty
		subtotal += it.Harga * it.Qty
		items = append(items, it)
	}

	// 2. Ambil semua promo non-voucher yang aktif saat ini di cabang terpilih
	promos, err := ambilPromoAktifCabang(ctx, tx, in.CabangID)
	if err != nil {
		return err
	}

	// 3. Validasi kode voucher (jika diisi)
	var voucherPromo *kalkulasi.Promo
	var voucherID int
	kodeVoucher := strings.TrimSpace(in.VoucherKode)
	if kodeVoucher != "" {
		vp, vid, err := validasiVoucherPOS(ctx, tx, kodeVoucher, in.CabangID, subtotal)
		if err != nil {
			return err
		}
		voucherPromo = vp
		voucherID = vid
	}

	// 4. Ambil tingkat member & saldo poin aktif (jika member dipilih)
	tingkat := "SILVER"
	saldoPoin := 0
	if in.MemberID > 0 {
		var aktif bool
		err := tx.QueryRowContext(ctx, `
			SELECT tingkat, aktif FROM member WHERE id = ?
		`, in.MemberID).Scan(&tingkat, &aktif)
		if errors.Is(err, sql.ErrNoRows) || !aktif {
			return ErrValidasi{Pesan: "Data member tidak ditemukan atau sedang nonaktif"}
		}
		if err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(sisa), 0)
			FROM poin_member
			WHERE member_id = ? AND status = 'AKTIF' AND sisa > 0 AND kadaluarsa_at > NOW()
		`, in.MemberID).Scan(&saldoPoin); err != nil {
			return err
		}
		if in.PoinDigunakan > saldoPoin {
			return ErrValidasi{Pesan: fmt.Sprintf("Poin yang ditukar (%d) melebihi saldo poin aktif member (%d poin)", in.PoinDigunakan, saldoPoin)}
		}
	} else {
		in.PoinDigunakan = 0
	}

	// 5. Jalankan kalkulasi promo kombinasi, aturan bentrok voucher, poin maks 50%, & reward poin baru
	hasil := kalkulasi.Hitung(items, promos, voucherPromo, saldoPoin, in.PoinDigunakan, tingkat)

	// Validasi jika kasir meminta tukar poin melebihi batas 50% total setelah promo
	maksPoin50 := (hasil.TotalSetelahPromo / 2) / kalkulasi.NilaiPoin
	if in.PoinDigunakan > maksPoin50 {
		return ErrValidasi{Pesan: fmt.Sprintf("Penukaran poin maksimal 50%% dari tagihan setelah promo (%d poin)", maksPoin50)}
	}

	// 6. Potong poin member dengan urutan FEFO
	poinTukar, err := tukarPoin(ctx, tx, in.MemberID, hasil.PoinDipakai)
	if err != nil {
		return err
	}

	var mID, vID any
	if in.MemberID > 0 {
		mID = in.MemberID
	}
	if voucherID > 0 && hasil.DiskonTransaksi > 0 {
		vID = voucherID
	}

	// 7. Simpan header transaksi
	res, err := tx.ExecContext(ctx, `
		INSERT INTO transaksi (cabang_id, member_id, voucher_id, poin_digunakan, subtotal, total_diskon, total_bayar, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'SELESAI')
	`, in.CabangID, mID, vID, poinTukar, hasil.Subtotal, hasil.TotalDiskon, hasil.TotalBayar)
	if err != nil {
		return err
	}
	transaksiID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// 8. Simpan detail transaksi_item & kurangi stok produk
	for _, it := range hasil.Items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transaksi_item (transaksi_id, produk_id, qty, harga_satuan, diskon_satuan, subtotal_item)
			VALUES (?, ?, ?, ?, ?, ?)
		`, transaksiID, it.ProdukID, it.Qty, it.Harga, it.DiskonSatuan, it.SubtotalItem); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE produk SET stok = stok - ? WHERE id = ?
		`, it.Qty, it.ProdukID); err != nil {
			return err
		}
	}

	// 9. Catat ledger transaksi_promo & naikkan counter promo.terpakai / voucher.terpakai
	for _, p := range hasil.Promo {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transaksi_promo (transaksi_id, promo_id, jenis, diskon)
			VALUES (?, ?, ?, ?)
		`, transaksiID, p.PromoID, p.Jenis, p.Diskon); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE promo SET terpakai = terpakai + 1 WHERE id = ?
		`, p.PromoID); err != nil {
			return err
		}
	}
	if voucherID > 0 && hasil.DiskonTransaksi > 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE voucher SET terpakai = terpakai + 1 WHERE id = ?
		`, voucherID); err != nil {
			return err
		}
	}

	// 10. Terbitkan poin baru dengan masa berlaku 6 bulan
	if in.MemberID > 0 && hasil.PoinDidapat > 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status)
			VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')
		`, in.MemberID, transaksiID, hasil.PoinDidapat, hasil.PoinDidapat); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ambilPromoAktifCabang(ctx context.Context, tx *sql.Tx, cabangID int) ([]kalkulasi.Promo, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT p.id, p.kode, p.jenis, p.nilai, p.minimal_belanja
		FROM promo p
		WHERE p.aktif = 1
		  AND p.jenis <> 'VOUCHER'
		  AND NOW() BETWEEN p.mulai AND p.berakhir
		  AND (p.mulai_jam IS NULL OR p.berakhir_jam IS NULL OR CURTIME() BETWEEN p.mulai_jam AND p.berakhir_jam)
		  AND (COALESCE(p.maksimal_pemakaian, 0) = 0 OR p.terpakai < p.maksimal_pemakaian)
		  AND (
		      NOT EXISTS (SELECT 1 FROM promo_cabang pc WHERE pc.promo_id = p.id)
		      OR EXISTS (SELECT 1 FROM promo_cabang pc WHERE pc.promo_id = p.id AND pc.cabang_id = ?)
		  )
	`, cabangID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []kalkulasi.Promo
	idx := make(map[int]int)
	for rows.Next() {
		var p kalkulasi.Promo
		if err := rows.Scan(&p.ID, &p.Kode, &p.Jenis, &p.Nilai, &p.MinimalBelanja); err != nil {
			return nil, err
		}
		p.ProdukIDs = make(map[int]bool)
		p.KategoriIDs = make(map[int]bool)
		idx[p.ID] = len(list)
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	if len(list) == 0 {
		return list, nil
	}

	rProd, err := tx.QueryContext(ctx, "SELECT promo_id, produk_id FROM promo_produk")
	if err != nil {
		return nil, err
	}
	for rProd.Next() {
		var pid, prid int
		if err := rProd.Scan(&pid, &prid); err == nil {
			if i, ok := idx[pid]; ok {
				list[i].ProdukIDs[prid] = true
			}
		}
	}
	rProd.Close()

	rKat, err := tx.QueryContext(ctx, "SELECT promo_id, kategori_id FROM promo_kategori")
	if err != nil {
		return nil, err
	}
	for rKat.Next() {
		var pid, kid int
		if err := rKat.Scan(&pid, &kid); err == nil {
			if i, ok := idx[pid]; ok {
				list[i].KategoriIDs[kid] = true
			}
		}
	}
	rKat.Close()

	return list, nil
}

func validasiVoucherPOS(ctx context.Context, tx *sql.Tx, kode string, cabangID, subtotal int) (*kalkulasi.Promo, int, error) {
	var vID, kuota, vTerpakai int
	var p kalkulasi.Promo
	var aktif, periodeOK, jamOK, kuotaPromoOK, cabangOK bool

	err := tx.QueryRowContext(ctx, `
		SELECT v.id, v.kuota, v.terpakai,
		       p.id, p.kode, p.jenis, p.nilai, p.minimal_belanja, p.aktif,
		       (NOW() BETWEEN p.mulai AND p.berakhir) AS periode_ok,
		       (p.mulai_jam IS NULL OR p.berakhir_jam IS NULL OR CURTIME() BETWEEN p.mulai_jam AND p.berakhir_jam) AS jam_ok,
		       (COALESCE(p.maksimal_pemakaian, 0) = 0 OR p.terpakai < p.maksimal_pemakaian) AS kuota_promo_ok,
		       (NOT EXISTS (SELECT 1 FROM promo_cabang pc WHERE pc.promo_id = p.id)
		        OR EXISTS (SELECT 1 FROM promo_cabang pc WHERE pc.promo_id = p.id AND pc.cabang_id = ?)) AS cabang_ok
		FROM voucher v
		JOIN promo p ON p.id = v.promo_id
		WHERE v.kode = ?
		FOR UPDATE
	`, cabangID, kode).Scan(
		&vID, &kuota, &vTerpakai,
		&p.ID, &p.Kode, &p.Jenis, &p.Nilai, &p.MinimalBelanja, &aktif,
		&periodeOK, &jamOK, &kuotaPromoOK, &cabangOK,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Kode voucher %q tidak ditemukan", kode)}
	}
	if err != nil {
		return nil, 0, err
	}
	if !aktif || !periodeOK {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Voucher %q sedang tidak aktif atau di luar periode berlaku", kode)}
	}
	if !jamOK {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Voucher %q hanya berlaku pada jam operasional promo", kode)}
	}
	if !cabangOK {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Voucher %q tidak berlaku di cabang ini", kode)}
	}
	if vTerpakai >= kuota || !kuotaPromoOK {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Kuota pemakaian voucher %q sudah habis", kode)}
	}
	if p.MinimalBelanja > 0 && subtotal < p.MinimalBelanja {
		return nil, 0, ErrValidasi{Pesan: fmt.Sprintf("Voucher %q membutuhkan minimal belanja Rp%d", kode, p.MinimalBelanja)}
	}
	p.Jenis = "VOUCHER"
	return &p, vID, nil
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
