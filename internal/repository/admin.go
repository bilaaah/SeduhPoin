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
    rows, err := r.DB.QueryContext(ctx, `
        SELECT
            v.id,
            v.promo_id,
            v.kode,
            v.kuota,
            v.terpakai,
            p.nama,
            p.jenis,
            p.nilai,
            p.minimal_belanja
        FROM voucher v
        JOIN promo p ON p.id = v.promo_id
        ORDER BY v.id DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.Voucher

    for rows.Next() {
        var x model.Voucher

        if err := rows.Scan(
            &x.ID,
            &x.PromoID,
            &x.Kode,
            &x.Kuota,
            &x.Terpakai,
            &x.PromoNama,
            &x.PromoJenis,
            &x.PromoNilai,
            &x.PromoMinimal,
        ); err != nil {
            return nil, err
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


func (r *AdminRepo) TambahTransaksiKasir(
	ctx context.Context,
	cabangID, memberID, voucherID, diskon, bayar, poinTukar, poinDapat int,
	items []model.TransaksiItemInput,
) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(items) == 0 {
		return errors.New("minimal satu produk harus dipilih")
	}

	if cabangID <= 0 || voucherID < 0 || diskon < 0 || bayar < 0 || poinTukar < 0 || poinDapat < 0 {
		return errors.New("data transaksi tidak valid")
	}

	type detailProduk struct {
		ProdukID int
		Jumlah   int
		Harga    int
	}

	type batchPoinDipakai struct {
		ID     int
		Jumlah int
	}

	var daftarProduk []detailProduk
	var daftarPoinDipakai []batchPoinDipakai

	// 1. Validasi produk, harga, dan stok.
	var subtotal int
	produkTerpilih := make(map[int]bool)

	for _, item := range items {
		if item.ProdukID <= 0 || item.Jumlah <= 0 {
			return errors.New("produk atau jumlah tidak valid")
		}

		if produkTerpilih[item.ProdukID] {
			return errors.New("produk yang sama tidak boleh dikirim dua kali")
		}
		produkTerpilih[item.ProdukID] = true

		var harga, stok int

		err := tx.QueryRowContext(ctx, `
			SELECT harga, stok
			FROM produk
			WHERE id = ?
			  AND aktif = 1
			  AND deleted_at IS NULL
			FOR UPDATE
		`, item.ProdukID).Scan(&harga, &stok)

		if err != nil {
			return err
		}

		if stok < item.Jumlah {
			return errors.New("stok produk tidak mencukupi")
		}

		subtotal += harga * item.Jumlah

		daftarProduk = append(daftarProduk, detailProduk{
			ProdukID: item.ProdukID,
			Jumlah:   item.Jumlah,
			Harga:    harga,
		})
	}

	// 2. Validasi diskon dan pembayaran.
	if voucherID > 0 {
		var (
			promoID       int
			promoJenis    string
			promoNilai    int
			promoMinimal  int
			promoMaksimal int
			promoTerpakai int
			voucherKuota  int
			voucherPakai  int
			promoAktif    bool
			promoBerlaku  bool
		)

		err := tx.QueryRowContext(ctx, `
			SELECT
				p.id,
				p.jenis,
				p.nilai,
				p.minimal_belanja,
				p.maksimal_pemakaian,
				p.terpakai,
				p.aktif,
				CASE
					WHEN NOW() BETWEEN p.mulai AND p.berakhir THEN 1
					ELSE 0
				END,
				v.kuota,
				v.terpakai
			FROM voucher v
			JOIN promo p ON p.id = v.promo_id
			WHERE v.id = ?
			FOR UPDATE
		`, voucherID).Scan(
			&promoID,
			&promoJenis,
			&promoNilai,
			&promoMinimal,
			&promoMaksimal,
			&promoTerpakai,
			&promoAktif,
			&promoBerlaku,
			&voucherKuota,
			&voucherPakai,
		)

		if err != nil {
			return errors.New("voucher tidak ditemukan")
		}

		if promoJenis != "VOUCHER" || !promoAktif || !promoBerlaku {
			return errors.New("voucher tidak aktif atau sudah kedaluwarsa")
		}

		if voucherPakai >= voucherKuota {
			return errors.New("kuota voucher sudah habis")
		}

		if promoMaksimal > 0 && promoTerpakai >= promoMaksimal {
			return errors.New("batas penggunaan promo sudah tercapai")
		}

		if subtotal < promoMinimal {
			return errors.New("subtotal belum memenuhi minimal penggunaan voucher")
		}

		if promoNilai < 0 || promoNilai > subtotal {
			return errors.New("nilai diskon voucher tidak valid")
		}

		// Nominal diskon diambil dari database, bukan dari form.
		diskon = promoNilai

		// Simpan ID promo untuk pencatatan pemakaian.
		_ = promoID
	} else {
		// Jika tidak menggunakan voucher, diskon manual tetap berlaku.
		if diskon > subtotal {
			return errors.New("diskon melebihi subtotal")
		}
	}

	if poinTukar > 0 && memberID <= 0 {
		return errors.New("pelanggan umum tidak dapat menukar poin")
	}

	if poinTukar*1000*2 > subtotal-diskon {
		return errors.New("penukaran poin melebihi batas 50 persen")
	}

	totalSeharusnya := subtotal - diskon - (poinTukar * 1000)

	if bayar != totalSeharusnya {
		return errors.New("total pembayaran tidak sesuai")
	}

	if memberID <= 0 && poinDapat > 0 {
		return errors.New("pelanggan umum tidak memperoleh poin")
	}

	// 3. Ambil batch poin berdasarkan FEFO.
	if poinTukar > 0 {
		rows, err := tx.QueryContext(ctx, `
			SELECT id, sisa
			FROM poin_member
			WHERE member_id = ?
			  AND status = 'AKTIF'
			  AND sisa > 0
			  AND kadaluarsa_at > NOW()
			ORDER BY kadaluarsa_at ASC
			FOR UPDATE
		`, memberID)

		if err != nil {
			return err
		}

		var batchPoin []batchPoinDipakai
		totalPoinTersedia := 0

		for rows.Next() {
			var id, sisa int

			if err := rows.Scan(&id, &sisa); err != nil {
				rows.Close()
				return err
			}

			batchPoin = append(batchPoin, batchPoinDipakai{
				ID:     id,
				Jumlah: sisa,
			})

			totalPoinTersedia += sisa
		}

		err = rows.Err()
		rows.Close()

		if err != nil {
			return err
		}

		if totalPoinTersedia < poinTukar {
			return errors.New("saldo poin member tidak mencukupi")
		}

		// Tentukan jumlah poin yang diambil dari setiap batch.
		sisaTukar := poinTukar

		for _, batch := range batchPoin {
			if sisaTukar <= 0 {
				break
			}

			pakai := batch.Jumlah
			if pakai > sisaTukar {
				pakai = sisaTukar
			}

			daftarPoinDipakai = append(daftarPoinDipakai, batchPoinDipakai{
				ID:     batch.ID,
				Jumlah: pakai,
			})

			sisaTukar -= pakai
		}
	}

	// 4. Simpan transaksi.
	var mID any
	var mVoucher any

	if memberID > 0 {
		mID = memberID
	}

	if voucherID > 0 {
		mVoucher = voucherID
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO transaksi (
			cabang_id,
			member_id,
			voucher_id,
			subtotal,
			total_diskon,
			total_bayar,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?, 'SELESAI')
	`, cabangID, mID, mVoucher, subtotal, diskon, bayar)

	if err != nil {
		return err
	}

	transaksiID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// 5. Catat penggunaan voucher.
	if voucherID > 0 {
		result, err := tx.ExecContext(ctx, `
			UPDATE voucher
			SET terpakai = terpakai + 1
			WHERE id = ?
			  AND terpakai < kuota
		`, voucherID)

		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if affected != 1 {
			return errors.New("kuota voucher sudah habis")
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE promo
			SET terpakai = terpakai + 1
			WHERE id = (
				SELECT promo_id
				FROM voucher
				WHERE id = ?
			)
		`, voucherID)

		if err != nil {
			return err
		}
	}

	// 6. Simpan detail produk dan kurangi stok.
	for _, item := range daftarProduk {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO transaksi_item (
				transaksi_id,
				produk_id,
				jumlah,
				harga_satuan
			)
			VALUES (?, ?, ?, ?)
		`, transaksiID, item.ProdukID, item.Jumlah, item.Harga)

		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE produk
			SET stok = stok - ?
			WHERE id = ?
		`, item.Jumlah, item.ProdukID)

		if err != nil {
			return err
		}
	}

	// 7. Kurangi saldo poin dan simpan riwayat penukaran.
	for _, p := range daftarPoinDipakai {
		result, err := tx.ExecContext(ctx, `
			UPDATE poin_member
			SET sisa = sisa - ?
			WHERE id = ?
			  AND sisa >= ?
			  AND status = 'AKTIF'
		`, p.Jumlah, p.ID, p.Jumlah)

		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if affected != 1 {
			return errors.New("gagal mengurangi saldo poin")
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO transaksi_poin_ditukar (
				transaksi_id,
				poin_member_id,
				jumlah_dipakai
			)
			VALUES (?, ?, ?)
		`, transaksiID, p.ID, p.Jumlah)

		if err != nil {
			return err
		}
	}

	// 8. Simpan poin reward yang diperoleh member.
	if memberID > 0 && poinDapat > 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO poin_member (
				member_id,
				transaksi_id,
				jumlah,
				sisa,
				diperoleh_at,
				kadaluarsa_at,
				status
			)
			VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')
		`, memberID, transaksiID, poinDapat, poinDapat)

		if err != nil {
			return err
		}
	}

	// 9. Simpan seluruh perubahan.
	return tx.Commit()
}
