package model

// AlokasiPromo adalah satu promo yang diterapkan pada sebuah transaksi
// beserta nominal rupiah diskon yang benar-benar diserap.
type AlokasiPromo struct {
	PromoID     int
	VoucherKode string
	Diskon      int
}

// TransaksiInput adalah payload lengkap penyimpanan transaksi kasir POS.
type TransaksiInput struct {
	CabangID      int
	MemberID      int
	VoucherID     int
	PoinDigunakan int
	Subtotal      int
	TotalDiskon   int
	TotalBayar    int
	PoinDiperoleh int
	Promo         []AlokasiPromo
}
