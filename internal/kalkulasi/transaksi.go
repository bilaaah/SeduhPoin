package kalkulasi

const (
	NilaiPoin     = 1000  // 1 poin = Rp1.000
	KelipatanPoin = 10000 // 1 poin per kelipatan Rp10.000
)

type Item struct {
	ProdukID   int
	KategoriID int
	Nama       string
	Harga      int
	Qty        int
}

type Promo struct {
	ID             int
	Kode           string
	Jenis          string // PERSENTASE_PRODUK, POTONGAN_MIN_BELANJA, BUY2_GET1, VOUCHER
	Nilai          int
	MinimalBelanja int
	ProdukIDs      map[int]bool
	KategoriIDs    map[int]bool
}

type HasilItem struct {
	Item
	DiskonTotal  int
	DiskonSatuan int
	SubtotalItem int
}

type PemakaianPromo struct {
	PromoID int
	Jenis   string
	Diskon  int
}

type Hasil struct {
	Items             []HasilItem
	Subtotal          int
	DiskonProduk      int
	DiskonTransaksi   int
	TotalDiskon       int
	TotalSetelahPromo int
	PoinDipakai       int
	PotonganPoin      int
	TotalBayar        int
	PoinDidapat       int
	Promo             []PemakaianPromo
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Hitung menghitung subtotal, diskon promo kombinasi, penukaran poin (maks 50%),
// total bayar akhir, serta perolehan poin baru (Gold 1.5x).
// Jika voucher != nil, promo POTONGAN_MIN_BELANJA tidak diterapkan (aturan bentrok).
func Hitung(items []Item, promos []Promo, voucher *Promo, saldoPoin, poinDiminta int, tingkat string) Hasil {
	var h Hasil
	pakai := make(map[int]*PemakaianPromo)
	var urutan []int

	catat := func(p Promo, d int) {
		if d <= 0 {
			return
		}
		if _, ok := pakai[p.ID]; !ok {
			pakai[p.ID] = &PemakaianPromo{PromoID: p.ID, Jenis: p.Jenis}
			urutan = append(urutan, p.ID)
		}
		pakai[p.ID].Diskon += d
	}

	// 1. Subtotal + diskon level produk (BUY2_GET1 & PERSENTASE_PRODUK)
	for _, it := range items {
		bruto := it.Harga * it.Qty
		h.Subtotal += bruto
		diskon := 0
		qtyBayar := it.Qty

		// BUY2_GET1: tiap kelipatan 3 item di kategori target, 1 item gratis
		for _, p := range promos {
			if p.Jenis == "BUY2_GET1" && p.KategoriIDs[it.KategoriID] {
				gratis := it.Qty / 3
				d := gratis * it.Harga
				diskon += d
				qtyBayar -= gratis
				catat(p, d)
				break
			}
		}

		// PERSENTASE_PRODUK: pilih persentase terbesar untuk produk target
		var terbaik *Promo
		for i := range promos {
			p := promos[i]
			if p.Jenis == "PERSENTASE_PRODUK" && p.ProdukIDs[it.ProdukID] &&
				(terbaik == nil || p.Nilai > terbaik.Nilai) {
				terbaik = &promos[i]
			}
		}
		if terbaik != nil {
			d := qtyBayar * it.Harga * terbaik.Nilai / 100
			diskon += d
			catat(*terbaik, d)
		}

		hi := HasilItem{
			Item:         it,
			DiskonTotal:  diskon,
			SubtotalItem: bruto - diskon,
		}
		if it.Qty > 0 {
			hi.DiskonSatuan = diskon / it.Qty
		}
		h.Items = append(h.Items, hi)
		h.DiskonProduk += diskon
	}

	setelahProduk := h.Subtotal - h.DiskonProduk

	// 2. Promo level transaksi: VOUCHER atau POTONGAN_MIN_BELANJA (tidak boleh digabung)
	if voucher != nil {
		d := minInt(voucher.Nilai, setelahProduk)
		h.DiskonTransaksi = d
		catat(*voucher, d)
	} else {
		var terbaik *Promo
		for i := range promos {
			p := promos[i]
			if p.Jenis == "POTONGAN_MIN_BELANJA" && h.Subtotal >= p.MinimalBelanja &&
				(terbaik == nil || p.Nilai > terbaik.Nilai) {
				terbaik = &promos[i]
			}
		}
		if terbaik != nil {
			d := minInt(terbaik.Nilai, setelahProduk)
			h.DiskonTransaksi = d
			catat(*terbaik, d)
		}
	}

	h.TotalDiskon = h.DiskonProduk + h.DiskonTransaksi
	h.TotalSetelahPromo = h.Subtotal - h.TotalDiskon

	// 3. Penukaran poin: maksimal 50% dari total belanja setelah promo
	if poinDiminta > 0 && saldoPoin > 0 {
		maksPoin := (h.TotalSetelahPromo / 2) / NilaiPoin
		h.PoinDipakai = minInt(poinDiminta, minInt(saldoPoin, maksPoin))
		h.PotonganPoin = h.PoinDipakai * NilaiPoin
	}
	h.TotalBayar = h.TotalSetelahPromo - h.PotonganPoin

	// 4. Perolehan poin baru: 1 poin per Rp10.000 (Gold 1.5x)
	poin := h.TotalBayar / KelipatanPoin
	if tingkat == "GOLD" {
		poin = poin * 3 / 2
	}
	h.PoinDidapat = poin

	for _, id := range urutan {
		h.Promo = append(h.Promo, *pakai[id])
	}
	return h
}
