package model

// FilterLaporan adalah parameter filter halaman Laporan Analitik.
// CabangID 0 berarti seluruh cabang, Periode kosong berarti seluruh periode.
type FilterLaporan struct {
	CabangID int
	Periode  string
}

// LaporanBiayaPromo adalah satu baris Laporan Biaya Realisasi Setiap Promo.
type LaporanBiayaPromo struct {
	Kode, Nama, Jenis string
	JumlahPemakaian   int
	TotalBiaya        int
	Persen            float64
}

// LaporanPoinMember adalah rincian saldo poin aktif yang dimiliki satu member.
type LaporanPoinMember struct {
	NomorMember         string
	Nama                string
	Tingkat             string
	JumlahBatch         int
	TotalPoin           int
	NilaiRupiah         int
	KedaluarsaTerdekat  string
}

// LaporanPoinCabang adalah sebaran liabilitas poin per cabang asal transaksi.
type LaporanPoinCabang struct {
	NamaCabang  string
	TotalPoin   int
	NilaiRupiah int
}

// LaporanTier adalah distribusi keanggotaan pada satu tingkat.
type LaporanTier struct {
	Tingkat string
	Jumlah  int
	Persen  float64
}

// LaporanPertumbuhan adalah pertumbuhan member baru pada satu bulan.
type LaporanPertumbuhan struct {
	Bulan          string // YYYY-MM
	LabelBulan     string // contoh: "Nov 2026"
	JumlahBaru     int
	TotalKumulatif int
	Persen         float64 // 0-100, tinggi batang grafik
}

// LaporanAnalitik adalah data lengkap untuk halaman /laporan.
type LaporanAnalitik struct {
	User   User
	Filter FilterLaporan

	Cabang          []Cabang
	PeriodeTersedia []string

	// Laporan 1: Biaya realisasi setiap promo
	ListPromo       []LaporanBiayaPromo
	TotalBiayaPromo int
	TotalPemakaian  int

	// Laporan 2: Poin yang masih berlaku
	PoinAktif      int
	NilaiPoin      int
	MemberBerpoin  int
	PoinPerMember  []LaporanPoinMember
	PoinPerCabang  []LaporanPoinCabang

	// Laporan 3: Perkembangan member per tingkat
	Tier            []LaporanTier
	TotalMember     int
	MemberGold      int
	MemberSilver    int
	Pertumbuhan     []LaporanPertumbuhan
	TotalMemberBaru int
}
