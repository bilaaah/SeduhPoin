-- SeduhPoin - schema siap import melalui phpMyAdmin/XAMPP
-- Catatan: database seduhpoin akan dibuat ulang agar schema lama tidak bentrok.

DROP DATABASE IF EXISTS seduhpoin;
CREATE DATABASE seduhpoin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE seduhpoin;

-- MASTER DATA
CREATE TABLE cabang (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    kode VARCHAR(20) NOT NULL,
    nama VARCHAR(100) NOT NULL,
    alamat VARCHAR(255) NOT NULL,
    telepon VARCHAR(30) NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id), UNIQUE KEY uk_cabang_kode (kode)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE kategori (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nama VARCHAR(80) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id), UNIQUE KEY uk_kategori_nama (nama)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE produk (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    kategori_id BIGINT UNSIGNED NULL,
    nama VARCHAR(100) NOT NULL,
    harga INT UNSIGNED NOT NULL DEFAULT 0,
    stok INT UNSIGNED NOT NULL DEFAULT 0,
    keterangan TEXT NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_produk_kategori (kategori_id),
    CONSTRAINT fk_produk_kategori FOREIGN KEY (kategori_id) REFERENCES kategori(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE member (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nomor_member VARCHAR(30) NOT NULL,
    nama VARCHAR(120) NOT NULL,
    no_hp VARCHAR(25) NOT NULL,
    email VARCHAR(150) NULL,
    alamat VARCHAR(255) NULL,
    tingkat ENUM('SILVER','GOLD') NOT NULL DEFAULT 'SILVER',
    total_belanja_3_bulan BIGINT UNSIGNED NOT NULL DEFAULT 0,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id), UNIQUE KEY uk_member_nomor (nomor_member), UNIQUE KEY uk_member_hp (no_hp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- PROMO DAN VOUCHER
CREATE TABLE promo (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    kode VARCHAR(40) NOT NULL,
    nama VARCHAR(120) NOT NULL,
    jenis ENUM('PERSENTASE_PRODUK','POTONGAN_MIN_BELANJA','BUY2_GET1','VOUCHER') NOT NULL,
    nilai INT UNSIGNED NOT NULL DEFAULT 0,
    minimal_belanja INT UNSIGNED NOT NULL DEFAULT 0,
    maksimal_pemakaian INT UNSIGNED NULL,
    terpakai INT UNSIGNED NOT NULL DEFAULT 0,
    mulai DATETIME NOT NULL,
    berakhir DATETIME NOT NULL,
    mulai_jam TIME NULL,
    berakhir_jam TIME NULL,
    aktif TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id), UNIQUE KEY uk_promo_kode (kode), KEY idx_promo_periode (mulai, berakhir)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_cabang (
    promo_id BIGINT UNSIGNED NOT NULL,
    cabang_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (promo_id, cabang_id),
    CONSTRAINT fk_promo_cabang_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE CASCADE,
    CONSTRAINT fk_promo_cabang_cabang FOREIGN KEY (cabang_id) REFERENCES cabang(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_produk (
    promo_id BIGINT UNSIGNED NOT NULL,
    produk_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (promo_id, produk_id),
    CONSTRAINT fk_promo_produk_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE CASCADE,
    CONSTRAINT fk_promo_produk_produk FOREIGN KEY (produk_id) REFERENCES produk(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_kategori (
    promo_id BIGINT UNSIGNED NOT NULL,
    kategori_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (promo_id, kategori_id),
    CONSTRAINT fk_promo_kategori_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE CASCADE,
    CONSTRAINT fk_promo_kategori_kategori FOREIGN KEY (kategori_id) REFERENCES kategori(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE voucher (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    promo_id BIGINT UNSIGNED NOT NULL,
    kode VARCHAR(50) NOT NULL,
    kuota INT UNSIGNED NOT NULL,
    terpakai INT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id), UNIQUE KEY uk_voucher_kode (kode), KEY idx_voucher_promo (promo_id),
    CONSTRAINT fk_voucher_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TABEL FONDASI MODUL TRANSAKSI, POIN, DAN LAPORAN
CREATE TABLE transaksi (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    cabang_id BIGINT UNSIGNED NOT NULL,
    member_id BIGINT UNSIGNED NULL,
    subtotal INT UNSIGNED NOT NULL DEFAULT 0,
    total_diskon INT UNSIGNED NOT NULL DEFAULT 0,
    total_bayar INT UNSIGNED NOT NULL DEFAULT 0,
    status ENUM('SELESAI','DIBATALKAN') NOT NULL DEFAULT 'SELESAI',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id), KEY idx_transaksi_cabang (cabang_id), KEY idx_transaksi_member (member_id),
    CONSTRAINT fk_transaksi_cabang FOREIGN KEY (cabang_id) REFERENCES cabang(id),
    CONSTRAINT fk_transaksi_member FOREIGN KEY (member_id) REFERENCES member(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE poin_member (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    member_id BIGINT UNSIGNED NOT NULL,
    transaksi_id BIGINT UNSIGNED NULL,
    jumlah INT NOT NULL,
    sisa INT NOT NULL,
    diperoleh_at DATETIME NOT NULL,
    kadaluarsa_at DATETIME NOT NULL,
    status ENUM('AKTIF','DIBATALKAN') NOT NULL DEFAULT 'AKTIF',
    PRIMARY KEY (id), KEY idx_poin_member (member_id), KEY idx_poin_expired (kadaluarsa_at),
    CONSTRAINT fk_poin_member FOREIGN KEY (member_id) REFERENCES member(id),
    CONSTRAINT fk_poin_transaksi FOREIGN KEY (transaksi_id) REFERENCES transaksi(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- DATA AWAL KEDAI KOPI SEDUH
INSERT INTO cabang (kode, nama, alamat, telepon) VALUES
('PUSAT', 'Seduh Pusat', 'Jl. Kopi No. 1', '021-111111'),
('TIMUR', 'Seduh Timur', 'Jl. Seduh Timur No. 2', '021-222222'),
('BARAT', 'Seduh Barat', 'Jl. Seduh Barat No. 3', '021-333333');

INSERT INTO kategori (nama) VALUES ('Kopi'), ('Non-Kopi'), ('Makanan');

INSERT INTO produk (kategori_id, nama, harga, stok, keterangan) VALUES
(1, 'Kopi Susu Gula Aren', 18000, 100, 'Kopi susu khas Kedai Seduh'),
(1, 'Americano', 15000, 100, 'Espresso dan air'),
(2, 'Matcha Latte', 22000, 50, 'Matcha dengan susu segar'),
(3, 'Croissant Butter', 20000, 40, 'Pastry panggang isi butter');
