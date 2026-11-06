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
	Aktif                                     bool
	CreatedAt                                 time.Time
}
type Promo struct {
	ID                                         int
	Kode, Nama, Jenis                          string
	Nilai, Minimal, Maksimal, Terpakai         int
	CabangID, TargetProdukID, TargetKategoriID int
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
}
