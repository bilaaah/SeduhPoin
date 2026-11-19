package kalkulasi

import "testing"

func TestGoldDapat15Poin(t *testing.T) {
	items := []Item{{ProdukID: 1, KategoriID: 1, Harga: 50000, Qty: 2}}
	h := Hitung(items, nil, nil, 0, 0, "GOLD")
	if h.TotalBayar != 100000 || h.PoinDidapat != 15 {
		t.Fatalf("dapat bayar=%d poin=%d, mau bayar=100000 poin=15", h.TotalBayar, h.PoinDidapat)
	}
}

func TestSilverDapat10Poin(t *testing.T) {
	items := []Item{{ProdukID: 1, KategoriID: 1, Harga: 50000, Qty: 2}}
	h := Hitung(items, nil, nil, 0, 0, "SILVER")
	if h.TotalBayar != 100000 || h.PoinDidapat != 10 {
		t.Fatalf("dapat bayar=%d poin=%d, mau bayar=100000 poin=10", h.TotalBayar, h.PoinDidapat)
	}
}

func TestBuy2Get1(t *testing.T) {
	p := Promo{ID: 3, Jenis: "BUY2_GET1", KategoriIDs: map[int]bool{1: true}}
	items := []Item{{ProdukID: 1, KategoriID: 1, Harga: 20000, Qty: 3}}
	h := Hitung(items, []Promo{p}, nil, 0, 0, "SILVER")
	if h.DiskonProduk != 20000 || h.TotalBayar != 40000 {
		t.Fatalf("diskon=%d bayar=%d, mau diskon=20000 bayar=40000", h.DiskonProduk, h.TotalBayar)
	}
}

func TestPersentaseProduk(t *testing.T) {
	p := Promo{ID: 1, Jenis: "PERSENTASE_PRODUK", Nilai: 20, ProdukIDs: map[int]bool{1: true}}
	items := []Item{{ProdukID: 1, KategoriID: 1, Harga: 25000, Qty: 2}}
	h := Hitung(items, []Promo{p}, nil, 0, 0, "SILVER")
	if h.DiskonProduk != 10000 || h.TotalBayar != 40000 {
		t.Fatalf("diskon=%d bayar=%d, mau diskon=10000 bayar=40000", h.DiskonProduk, h.TotalBayar)
	}
}

func TestVoucherMematikanMinBelanja(t *testing.T) {
	min := Promo{ID: 2, Jenis: "POTONGAN_MIN_BELANJA", Nilai: 15000, MinimalBelanja: 50000}
	v := Promo{ID: 4, Jenis: "VOUCHER", Nilai: 5000}
	items := []Item{{ProdukID: 1, Harga: 60000, Qty: 1}}
	h := Hitung(items, []Promo{min}, &v, 0, 0, "SILVER")
	if h.DiskonTransaksi != 5000 {
		t.Fatalf("harusnya cuma voucher 5000, dapat %d", h.DiskonTransaksi)
	}
}

func TestPoinMaks50Persen(t *testing.T) {
	items := []Item{{ProdukID: 1, Harga: 30000, Qty: 1}}
	h := Hitung(items, nil, nil, 100, 100, "SILVER")
	if h.PoinDipakai != 15 || h.TotalBayar != 15000 {
		t.Fatalf("poin=%d bayar=%d, mau poin=15 bayar=15000", h.PoinDipakai, h.TotalBayar)
	}
}
