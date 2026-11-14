package model

import "time"

// BatchPoin = satu baris poin_member beserta status turunannya.
// Status: AKTIF, KEDALUWARSA (kadaluarsa_at <= NOW), atau DIBATALKAN.
type BatchPoin struct {
	ID           int
	TransaksiID  int // 0 = poin warisan tanpa transaksi
	Jumlah, Sisa int
	DiperolehAt  time.Time
	KadaluarsaAt time.Time
	Status       string
	SisaHari     int // hari menuju kedaluwarsa (negatif jika sudah lewat)
}

// MemberTier = ringkasan loyalitas satu member.
type MemberTier struct {
	ID                   int
	Nomor, Nama          string
	TingkatTersimpan     string // nilai kolom member.tingkat saat ini
	TingkatSeharusnya    string // hasil evaluasi belanja riil 3 bulan
	BelanjaTersimpan     int    // member.total_belanja_3_bulan
	BelanjaRiil          int    // dihitung langsung dari transaksi SELESAI
	PoinAktif            int
	NilaiRupiah          int // PoinAktif * 1000
	KurangKeGold         int // 0 jika sudah memenuhi syarat Gold
	PerluSinkron         bool
}

const (
	BatasGold       = 1500000
	NilaiTukarPoin  = 1000
)