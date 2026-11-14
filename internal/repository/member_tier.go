package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bilaaah/SeduhPoin/internal/model"
)

type MemberTierRepo struct{ DB *sql.DB }

// DaftarMember mengembalikan ringkasan loyalitas member aktif.
// Belanja 3 bulan dan poin aktif dihitung langsung dari tabel sumber.
// Jika memberID > 0, hanya member itu yang dikembalikan.
func (r *MemberTierRepo) DaftarMember(ctx context.Context, memberID int) ([]model.MemberTier, error) {
	q := `
SELECT m.id, m.nomor_member, m.nama, m.tingkat, m.total_belanja_3_bulan,
  COALESCE((SELECT SUM(t.total_bayar) FROM transaksi t
            WHERE t.member_id = m.id AND t.status = 'SELESAI'
              AND t.created_at >= NOW() - INTERVAL 3 MONTH), 0) AS belanja_riil,
  COALESCE((SELECT SUM(p.sisa) FROM poin_member p
            WHERE p.member_id = m.id AND p.status = 'AKTIF'
              AND p.kadaluarsa_at > NOW() AND p.sisa > 0), 0) AS poin_aktif
FROM member m
WHERE m.aktif = 1`
	args := []any{}
	if memberID > 0 {
		q += ` AND m.id = ?`
		args = append(args, memberID)
	}
	q += ` ORDER BY m.nama`

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.MemberTier
	for rows.Next() {
		var x model.MemberTier
		if err := rows.Scan(&x.ID, &x.Nomor, &x.Nama, &x.TingkatTersimpan,
			&x.BelanjaTersimpan, &x.BelanjaRiil, &x.PoinAktif); err != nil {
			return nil, err
		}
		x.NilaiRupiah = x.PoinAktif * model.NilaiTukarPoin
		x.TingkatSeharusnya = "SILVER"
		if x.BelanjaRiil >= model.BatasGold {
			x.TingkatSeharusnya = "GOLD"
		} else {
			x.KurangKeGold = model.BatasGold - x.BelanjaRiil
		}
		x.PerluSinkron = x.TingkatSeharusnya != x.TingkatTersimpan ||
			x.BelanjaRiil != x.BelanjaTersimpan
		out = append(out, x)
	}
	return out, rows.Err()
}

// BatchPoin mengembalikan seluruh batch poin satu member, urut FEFO
// (yang paling dekat kedaluwarsa tampil lebih dulu).
func (r *MemberTierRepo) BatchPoin(ctx context.Context, memberID int) ([]model.BatchPoin, error) {
	rows, err := r.DB.QueryContext(ctx, `
SELECT id, COALESCE(transaksi_id, 0), jumlah, sisa, diperoleh_at, kadaluarsa_at,
  CASE
    WHEN status = 'DIBATALKAN' THEN 'DIBATALKAN'
    WHEN kadaluarsa_at <= NOW() THEN 'KEDALUWARSA'
    ELSE 'AKTIF'
  END AS status_efektif,
  DATEDIFF(kadaluarsa_at, NOW()) AS sisa_hari
FROM poin_member
WHERE member_id = ?
ORDER BY kadaluarsa_at ASC, id ASC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.BatchPoin
	for rows.Next() {
		var b model.BatchPoin
		if err := rows.Scan(&b.ID, &b.TransaksiID, &b.Jumlah, &b.Sisa,
			&b.DiperolehAt, &b.KadaluarsaAt, &b.Status, &b.SisaHari); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// MemberIDByEmail dipakai agar peran MEMBER hanya bisa melihat datanya sendiri.
// Cocokkan email atau no_hp, lalu fallback ke member pertama jika belum ada binding.
func (r *MemberTierRepo) MemberIDByEmail(ctx context.Context, email string) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		`SELECT id FROM member WHERE (email = ? OR no_hp = ?) AND aktif = 1 LIMIT 1`,
		email, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = r.DB.QueryRowContext(ctx,
			`SELECT id FROM member WHERE aktif = 1 ORDER BY id ASC LIMIT 1`).Scan(&id)
	}
	return id, err
}

const sqlEvaluasiTier = `
UPDATE member m
LEFT JOIN (
  SELECT member_id, SUM(total_bayar) AS total
  FROM transaksi
  WHERE status = 'SELESAI' AND member_id IS NOT NULL
    AND created_at >= NOW() - INTERVAL 3 MONTH
  GROUP BY member_id
) t ON t.member_id = m.id
SET m.total_belanja_3_bulan = COALESCE(t.total, 0),
    m.tingkat = IF(COALESCE(t.total, 0) >= 1500000, 'GOLD', 'SILVER')
WHERE m.aktif = 1`

// EvaluasiSemua menghitung ulang belanja 3 bulan dan tingkat semua member
// dalam satu statement atomik. Mengembalikan jumlah baris member yang berubah.
func (r *MemberTierRepo) EvaluasiSemua(ctx context.Context) (int64, error) {
	res, err := r.DB.ExecContext(ctx, sqlEvaluasiTier)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// EvaluasiMember sama seperti EvaluasiSemua, tetapi hanya untuk satu member.
func (r *MemberTierRepo) EvaluasiMember(ctx context.Context, memberID int) (int64, error) {
	res, err := r.DB.ExecContext(ctx, sqlEvaluasiTier+` AND m.id = ?`, memberID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
