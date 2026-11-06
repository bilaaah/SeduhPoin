-- SeduhPoin - schema siap import melalui phpMyAdmin/XAMPP
-- Catatan: database seduhpoin akan dibuat ulang agar schema lama tidak bentrok.

DROP DATABASE IF EXISTS seduhpoin;
CREATE DATABASE seduhpoin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE seduhpoin;

-- PENGGUNA DAN AUTENTIKASI (Session & Multi-Role)
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('ADMIN', 'KASIR', 'MEMBER') NOT NULL DEFAULT 'ADMIN',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
    voucher_id BIGINT UNSIGNED NULL,         -- voucher yang dipakai (jika ada), untuk restock oleh Hilal
    poin_digunakan INT UNSIGNED NOT NULL DEFAULT 0, -- total poin yang ditukarkan di transaksi ini
    subtotal INT UNSIGNED NOT NULL DEFAULT 0,
    total_diskon INT UNSIGNED NOT NULL DEFAULT 0,
    total_bayar INT UNSIGNED NOT NULL DEFAULT 0,
    status ENUM('SELESAI','DIBATALKAN') NOT NULL DEFAULT 'SELESAI',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_transaksi_cabang (cabang_id),
    KEY idx_transaksi_member (member_id),
    CONSTRAINT fk_transaksi_cabang FOREIGN KEY (cabang_id) REFERENCES cabang(id),
    CONSTRAINT fk_transaksi_member FOREIGN KEY (member_id) REFERENCES member(id) ON DELETE SET NULL,
    CONSTRAINT fk_transaksi_voucher FOREIGN KEY (voucher_id) REFERENCES voucher(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Detail item per transaksi (untuk kalkulasi promo per produk & restock saat batal)
CREATE TABLE transaksi_item (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    transaksi_id BIGINT UNSIGNED NOT NULL,
    produk_id BIGINT UNSIGNED NOT NULL,
    qty INT UNSIGNED NOT NULL DEFAULT 1,
    harga_satuan INT UNSIGNED NOT NULL,       -- harga saat transaksi (snapshot)
    diskon_satuan INT UNSIGNED NOT NULL DEFAULT 0, -- diskon per unit dari promo produk/B2G1
    subtotal_item INT UNSIGNED NOT NULL,      -- (harga_satuan - diskon_satuan) * qty
    PRIMARY KEY (id),
    KEY idx_item_transaksi (transaksi_id),
    KEY idx_item_produk (produk_id),
    CONSTRAINT fk_item_transaksi FOREIGN KEY (transaksi_id) REFERENCES transaksi(id) ON DELETE CASCADE,
    CONSTRAINT fk_item_produk FOREIGN KEY (produk_id) REFERENCES produk(id)
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

-- DATA SEED MEMBER UNTUK SIMULASI (Lingkup 1 & 5)
INSERT INTO member (id, nomor_member, nama, no_hp, email, alamat, tingkat, total_belanja_3_bulan) VALUES
(1, 'MBR-001', 'Nabila Zahra', '081234567890', 'nabila@gmail.com', 'Jl. Melati No. 12, Jakarta', 'GOLD', 1850000),
(2, 'MBR-002', 'Budi Santoso', '081298765432', 'budi.santoso@yahoo.com', 'Jl. Cempaka No. 5, Jakarta', 'SILVER', 450000),
(3, 'MBR-003', 'Siti Rahmawati', '081377889900', 'siti.rahma@gmail.com', 'Jl. Mawar No. 8, Jakarta', 'GOLD', 2100000),
(4, 'MBR-004', 'Dimas Pratama', '085611223344', 'dimas@outlook.com', 'Jl. Kenanga No. 19, Jakarta', 'SILVER', 0);

-- DATA SEED POIN MEMBER (Masa berlaku 6 bulan & simulasi FEFO)
INSERT INTO poin_member (member_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status) VALUES
(1, 150, 150, '2026-08-01 10:00:00', '2027-02-01 10:00:00', 'AKTIF'),
(1, 200, 200, '2026-09-15 14:30:00', '2027-03-15 14:30:00', 'AKTIF'),
(2, 45, 45, '2026-10-05 11:20:00', '2027-04-05 11:20:00', 'AKTIF'),
(3, 300, 300, '2026-07-20 09:15:00', '2027-01-20 09:15:00', 'AKTIF');

-- DATA SEED PROMO (Lingkup 2)
-- 1. Promo Diskon Persentase Produk 20%
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(1, 'DISKONKOPI20', 'Diskon Kopi 20% Spesial Seduh', 'PERSENTASE_PRODUK', 20, 0, 100, 5, '2026-01-01 00:00:00', '2026-12-31 23:59:59', 1);

INSERT INTO promo_produk (promo_id, produk_id) VALUES
(1, 1), -- Kopi Susu Gula Aren
(1, 2); -- Americano

-- 2. Promo Potongan Minimal Belanja Rp15.000 Min. Rp50.000
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(2, 'HEMAT50K', 'Potongan Belanja Rp15.000 Min. Rp50.000', 'POTONGAN_MIN_BELANJA', 15000, 50000, 200, 12, '2026-01-01 00:00:00', '2026-12-31 23:59:59', 1);

-- 3. Promo Beli 2 Gratis 1 Kategori Kopi (Happy Hour 14:00 - 17:00 di Cabang Pusat & Timur)
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, mulai_jam, berakhir_jam, aktif) VALUES
(3, 'B2G1KOPI', 'Happy Hour Beli 2 Gratis 1 Kategori Kopi', 'BUY2_GET1', 0, 0, 50, 8, '2026-01-01 00:00:00', '2026-12-31 23:59:59', '14:00:00', '17:00:00', 1);

INSERT INTO promo_kategori (promo_id, kategori_id) VALUES
(3, 1); -- Kategori Kopi

INSERT INTO promo_cabang (promo_id, cabang_id) VALUES
(3, 1), -- Cabang Pusat
(3, 2); -- Cabang Timur

-- 4. Promo Voucher Potongan Rp10.000
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(4, 'PROMO-VOUCHER-10K', 'Promo Voucher Potongan Rp10.000', 'VOUCHER', 10000, 30000, 150, 25, '2026-01-01 00:00:00', '2026-12-31 23:59:59', 1);

-- 5. Promo Voucher Akhir Pekan Rp20.000 (Khusus Cabang Barat)
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(5, 'PROMO-VOUCHER-20K', 'Promo Voucher Akhir Pekan Rp20.000 Cabang Barat', 'VOUCHER', 20000, 60000, 50, 3, '2026-01-01 00:00:00', '2026-12-31 23:59:59', 1);

INSERT INTO promo_cabang (promo_id, cabang_id) VALUES
(5, 3); -- Cabang Barat

-- DATA SEED VOUCHER BERKODE (Lingkup 2)
INSERT INTO voucher (promo_id, kode, kuota, terpakai) VALUES
(4, 'SEDUHSERU10', 50, 15),
(4, 'KOPIHEMAT10', 50, 10),
(4, 'WELCOMEGIFT', 50, 0),
(5, 'WEEKENDSEDUH', 50, 3);

-- DATA SEED USER SIMULASI (Password: admin123, kasir123, member123)
INSERT INTO users (id, nama, email, password_hash, role) VALUES
(1, 'Pemilik Kedai (Admin)', 'admin@seduhkopi.com', '$2a$10$KPQoj7P.EAmlSLoSAAfFtuMF42IOayaW2jtlLMMGacBwFja/xk6Me', 'ADMIN'),
(2, 'Kasir Kedai Seduh', 'kasir@seduhkopi.com', '$2a$10$sPRcFmGKAwipGlTTJ/62UO7NRWEfB108zfbB34wKqfQ61hDEre6N6', 'KASIR'),
(3, 'Budi (Member Loyal)', 'member@seduhkopi.com', '$2a$10$qIbJFbrmTU.sRWp7JC5RiO6bOvntjkPlUf9Iswa/PtiKKl.RUMox6', 'MEMBER');
