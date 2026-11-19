package model

import "time"

type Kategori struct {
	ID   int
	Nama string
}

type Produk struct {
	ID         int
	KategoriID int
	Kategori   string
	Nama       string
	Harga      int
	Stok       int
	Keterangan string
	Aktif      bool
}

type Cabang struct {
	ID                          int
	Kode, Nama, Alamat, Telepon string
	Aktif                       bool
}

type Member struct {
	ID                                        int
	Nomor, Nama, NoHP, Email, Alamat, Tingkat string
	Total                                     int
	SaldoPoin                                 int
	Aktif                                     bool
	CreatedAt                                 time.Time
}

type Promo struct {
	ID                                         int
	Kode, Nama, Jenis                          string
	Nilai, Minimal, Maksimal, Terpakai         int
	CabangID, TargetProdukID, TargetKategoriID int
	CabangIDs, ProdukIDs, KategoriIDs          []int
	MulaiJam, BerakhirJam                      string
	Mulai, Berakhir                            time.Time
	Aktif                                      bool
}

type Voucher struct {
	ID, PromoID, Kuota, Terpakai int
	Kode, PromoNama              string
}

type RingkasanData struct {
	User         User
	TotalProduk  int
	TotalCabang  int
	TotalMember  int
	MemberGold   int
	MemberSilver int
	TotalPromo   int
	TotalVoucher int

	// Untuk Kasir
	KasirData KasirData

	// Untuk Member
	MemberData MemberPortalData
}

type Transaksi struct {
	ID            int
	CabangID      int
	CabangNama    string
	MemberID      int
	MemberNama    string
	MemberNomor   string
	Subtotal      int
	TotalDiskon   int
	TotalBayar    int
	PoinDitukar   int
	PoinDiperoleh int
	Status        string
	CreatedAt     time.Time
}

type PoinDetail struct {
	ID           int
	MemberID     int
	Jumlah       int
	Sisa         int
	DiperolehAt  time.Time
	KadaluarsaAt time.Time
	Status       string
}

type KasirData struct {
	User      User
	Cabang    []Cabang
	Produk    []Produk
	Member    []Member
	Promo     []Promo
	Voucher   []Voucher
	Transaksi []Transaksi
}

type MemberPortalData struct {
	User           User
	Member         Member
	PoinAktif      int
	NilaiRupiah    int
	SisaMenujuGold int
	BatchPoin      []PoinDetail
	Riwayat        []Transaksi
}
