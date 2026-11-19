package model

// AlokasiPromo adalah satu promo yang diterapkan pada sebuah transaksi
// beserta nominal rupiah diskon yang benar-benar diserap.
type AlokasiPromo struct {
	PromoID     int
	VoucherKode string
	Diskon      int
}

// ItemBelanja adalah satu baris produk dan kuantitas yang dipilih kasir di POS.
type ItemBelanja struct {
	ProdukID int
	Qty      int
}

// TransaksiInput adalah payload lengkap penyimpanan transaksi kasir POS.
type TransaksiInput struct {
	CabangID      int
	MemberID      int
	VoucherID     int
	VoucherKode   string
	PoinDigunakan int
	Subtotal      int
	TotalDiskon   int
	TotalBayar    int
	PoinDiperoleh int
	Items         []ItemBelanja
	Promo         []AlokasiPromo
}
