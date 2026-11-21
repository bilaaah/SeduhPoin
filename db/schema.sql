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
    role ENUM('ADMIN', 'KASIR', 'MEMBER') NOT NULL DEFAULT 'MEMBER',
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

-- LEDGER PEMAKAIAN PROMO (Lingkup 6)
-- Satu baris = satu promo yang benar-benar dipakai pada satu transaksi, beserta
-- nominal rupiah diskon yang diserap. Inilah sumber kebenaran "biaya promo"
-- pada Laporan Analitik, bukan lagi estimasi dari kolom promo.terpakai.
CREATE TABLE transaksi_promo (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    transaksi_id BIGINT UNSIGNED NOT NULL,
    promo_id BIGINT UNSIGNED NOT NULL,
    jenis ENUM('PERSENTASE_PRODUK','POTONGAN_MIN_BELANJA','BUY2_GET1','VOUCHER') NOT NULL,
    diskon INT UNSIGNED NOT NULL DEFAULT 0,   -- rupiah yang benar-benar dipotong pada transaksi ini
    PRIMARY KEY (id),
    UNIQUE KEY uk_transaksi_promo (transaksi_id, promo_id),
    KEY idx_trxpromo_promo (promo_id),
    CONSTRAINT fk_trxpromo_transaksi FOREIGN KEY (transaksi_id) REFERENCES transaksi(id) ON DELETE CASCADE,
    CONSTRAINT fk_trxpromo_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Index untuk query pertumbuhan member per bulan pada Laporan Analitik
ALTER TABLE member ADD KEY idx_member_created (created_at);

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
-- created_at di-back-date relatif terhadap NOW() supaya Laporan Perkembangan
-- Member per Tingkat selalu punya riwayat 4 bulan terakhir, berapa pun tanggal
-- data ini diimpor. Nilai relative dipakai karena kolom bertipe TIMESTAMP
-- (batas 2038) dan jendela laporan "12 bulan terakhir" selalu bergeser.
-- Catatan: total_belanja_3_bulan mencerminkan akumulasi belanja riil kedai;
-- transaksi simulasi di bawah hanya sampel, bukan keseluruhan riwayat.
INSERT INTO member (id, nomor_member, nama, no_hp, email, alamat, tingkat, total_belanja_3_bulan, created_at) VALUES
(1, 'MBR-001', 'Nabila Zahra', '081234567890', 'nabila@gmail.com', 'Jl. Melati No. 12, Jakarta', 'GOLD', 1850000, DATE_SUB(NOW(), INTERVAL 8 MONTH)),
(2, 'MBR-002', 'Budi Santoso', '081298765432', 'member@seduhkopi.com', 'Jl. Cempaka No. 5, Jakarta', 'SILVER', 125000, DATE_SUB(NOW(), INTERVAL 6 MONTH)),
(3, 'MBR-003', 'Siti Rahmawati', '081377889900', 'siti.rahma@gmail.com', 'Jl. Mawar No. 8, Jakarta', 'GOLD', 2100000, DATE_SUB(NOW(), INTERVAL 4 MONTH)),
(4, 'MBR-004', 'Dimas Pratama', '085611223344', 'dimas@outlook.com', 'Jl. Kenanga No. 19, Jakarta', 'SILVER', 150000, DATE_SUB(NOW(), INTERVAL 3 MONTH));

-- DATA SEED PROMO (Lingkup 2)
-- Catatan: kolom `terpakai` diselaraskan dengan jumlah pemakaian nyata pada
-- tabel transaksi_promo di bawah, sehingga Laporan Biaya Promo konsisten.
-- 1. Promo Diskon Persentase Produk 20%
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(1, 'DISKONKOPI20', 'Diskon Kopi 20% Spesial Seduh', 'PERSENTASE_PRODUK', 20, 0, 100, 3, DATE_SUB(NOW(), INTERVAL 6 MONTH), DATE_ADD(NOW(), INTERVAL 6 MONTH), 1);

INSERT INTO promo_produk (promo_id, produk_id) VALUES
(1, 1), -- Kopi Susu Gula Aren
(1, 2); -- Americano

-- 2. Promo Potongan Minimal Belanja Rp15.000 Min. Rp50.000
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(2, 'HEMAT50K', 'Potongan Belanja Rp15.000 Min. Rp50.000', 'POTONGAN_MIN_BELANJA', 15000, 50000, 200, 2, DATE_SUB(NOW(), INTERVAL 6 MONTH), DATE_ADD(NOW(), INTERVAL 6 MONTH), 1);

-- 3. Promo Beli 2 Gratis 1 Kategori Kopi (Happy Hour 14:00 - 17:00 di Cabang Pusat & Timur)
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, mulai_jam, berakhir_jam, aktif) VALUES
(3, 'B2G1KOPI', 'Happy Hour Beli 2 Gratis 1 Kategori Kopi', 'BUY2_GET1', 0, 0, 50, 1, DATE_SUB(NOW(), INTERVAL 6 MONTH), DATE_ADD(NOW(), INTERVAL 6 MONTH), '14:00:00', '17:00:00', 1);

INSERT INTO promo_kategori (promo_id, kategori_id) VALUES
(3, 1); -- Kategori Kopi

INSERT INTO promo_cabang (promo_id, cabang_id) VALUES
(3, 1), -- Cabang Pusat
(3, 2); -- Cabang Timur

-- 4. Promo Voucher Potongan Rp10.000
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(4, 'PROMO-VOUCHER-10K', 'Promo Voucher Potongan Rp10.000', 'VOUCHER', 10000, 30000, 150, 3, DATE_SUB(NOW(), INTERVAL 6 MONTH), DATE_ADD(NOW(), INTERVAL 6 MONTH), 1);

-- 5. Promo Voucher Akhir Pekan Rp20.000 (Khusus Cabang Barat)
INSERT INTO promo (id, kode, nama, jenis, nilai, minimal_belanja, maksimal_pemakaian, terpakai, mulai, berakhir, aktif) VALUES
(5, 'PROMO-VOUCHER-20K', 'Promo Voucher Akhir Pekan Rp20.000 Cabang Barat', 'VOUCHER', 20000, 60000, 50, 2, DATE_SUB(NOW(), INTERVAL 6 MONTH), DATE_ADD(NOW(), INTERVAL 6 MONTH), 1);

INSERT INTO promo_cabang (promo_id, cabang_id) VALUES
(5, 3); -- Cabang Barat

-- DATA SEED VOUCHER BERKODE (Lingkup 2)
-- `terpakai` diselaraskan dengan ledger transaksi_promo di bawah: 3 pemakaian
-- promo voucher Rp10.000 (1 per kode) dan 2 pemakaian voucher Rp20.000.
INSERT INTO voucher (promo_id, kode, kuota, terpakai) VALUES
(4, 'SEDUHSERU10', 50, 1),
(4, 'KOPIHEMAT10', 50, 1),
(4, 'WELCOMEGIFT', 50, 1),
(5, 'WEEKENDSEDUH', 50, 2);

-- DATA SEED TRANSAKSI SIMULASI (Lingkup 3 & 6)
-- 13 transaksi tersebar di 3 cabang dalam ~4 bulan terakhir. created_at memakai
-- ekspresi relatif (DATE_SUB NOW) supaya data simulasi selalu berada di dalam
-- jendela "12 bulan terakhir" hasil laporan, kapan pun schema.sql diimpor.
-- T13 berstatus DIBATALKAN sebagai bukti laporan biaya promo & poin tidak
-- menghitung transaksi yang dibatalkan.
INSERT INTO transaksi (id, cabang_id, member_id, voucher_id, poin_digunakan, subtotal, total_diskon, total_bayar, status, created_at) VALUES
(1,  1, 1, NULL,  0,  54000, 10800,  43200, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 118 DAY)),
(2,  2, 2, NULL,  0,  84000, 15000,  69000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 110 DAY)),
(3,  1, 3, 1,    0,  80000, 10000,  70000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 96 DAY)),
(4,  3, 4, 4,    0,  80000, 20000,  60000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 84 DAY)),
(5,  2, 1, NULL,  0,  84000, 18000,  66000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 72 DAY)),
(6,  1, 2, NULL,  0,  45000,     0,  45000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 62 DAY)),
(7,  3, 3, NULL,  0, 102000,  7200,  94800, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 50 DAY)),
(8,  2, 4, 2,    5,  70000, 10000,  55000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 40 DAY)),
(9,  1, 1, NULL, 12, 124000, 15000,  97000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 30 DAY)),
(10, 2, 3, 3,    0,  62000, 10000,  52000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 20 DAY)),
(11, 1, 2, NULL,  0,  80000,  7200,  72800, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 12 DAY)),
(12, 3, 1, 4,   20, 108000, 20000,  68000, 'SELESAI',    DATE_SUB(NOW(), INTERVAL 5 DAY)),
(13, 2, 3, NULL,  0,  80000,  7200,  72800, 'DIBATALKAN', DATE_SUB(NOW(), INTERVAL 2 DAY));

-- Detail item: subtotal_item = (harga_satuan - diskon_satuan) * qty.
-- SUM(harga_satuan * qty) = transaksi.subtotal dan SUM(subtotal_item) =
-- transaksi.subtotal - transaksi.total_diskon.
INSERT INTO transaksi_item (transaksi_id, produk_id, qty, harga_satuan, diskon_satuan, subtotal_item) VALUES
(1,  1, 3, 18000, 3600, 43200),
(2,  3, 2, 22000, 7500, 29000),
(2,  4, 2, 20000,    0, 40000),
(3,  1, 2, 18000,    0, 36000),
(3,  3, 2, 22000, 5000, 34000),
(4,  3, 2, 22000,10000, 24000),
(4,  1, 2, 18000,    0, 36000),
(5,  1, 3, 18000, 6000, 36000),
(5,  2, 2, 15000,    0, 30000),
(6,  2, 3, 15000,    0, 45000),
(7,  3, 3, 22000,    0, 66000),
(7,  1, 2, 18000, 3600, 28800),
(8,  4, 2, 20000, 5000, 30000),
(8,  2, 2, 15000,    0, 30000),
(9,  3, 1, 22000,    0, 22000),
(9,  1, 4, 18000, 3750, 57000),
(9,  2, 2, 15000,    0, 30000),
(10, 3, 2, 22000, 5000, 34000),
(10, 1, 1, 18000,    0, 18000),
(11, 3, 2, 22000,    0, 44000),
(11, 1, 2, 18000, 3600, 28800),
(12, 3, 1, 22000, 5000, 17000),
(12, 1, 2, 18000,    0, 36000),
(12, 2, 2, 15000,    0, 30000),
(12, 4, 1, 20000, 5000, 15000),
(13, 3, 2, 22000,    0, 44000),
(13, 1, 2, 18000, 3600, 28800);

-- LEDGER PEMAKAIAN PROMO (sumber kebenaran biaya promo pada Laporan Analitik).
-- Total biaya promo untuk transaksi SELESAI = Rp143.200:
--   DISKONKOPI20 (P1) 3x Rp25.200 | HEMAT50K (P2) 2x Rp30.000
--   B2G1KOPI  (P3) 1x Rp18.000 | PROMO-VOUCHER-10K (P4) 3x Rp30.000
--   PROMO-VOUCHER-20K (P5) 2x Rp40.000
-- Baris T13 tidak boleh ikut terhitung karena transaksinya DIBATALKAN.
INSERT INTO transaksi_promo (transaksi_id, promo_id, jenis, diskon) VALUES
(1,  1, 'PERSENTASE_PRODUK',     10800),
(2,  2, 'POTONGAN_MIN_BELANJA',   15000),
(3,  4, 'VOUCHER',                10000),
(4,  5, 'VOUCHER',                20000),
(5,  3, 'BUY2_GET1',             18000),
(7,  1, 'PERSENTASE_PRODUK',      7200),
(8,  4, 'VOUCHER',                10000),
(9,  2, 'POTONGAN_MIN_BELANJA',   15000),
(10, 4, 'VOUCHER',                10000),
(11, 1, 'PERSENTASE_PRODUK',      7200),
(12, 5, 'VOUCHER',                20000),
(13, 1, 'PERSENTASE_PRODUK',      7200);

-- DATA SEED POIN MEMBER (Lingkup 5 & 6)
-- 4 batch warisan (transaksi_id NULL) + 1 batch per transaksi. Poin yang
-- ditukar memakai FEFO: konsumsi diambil dari batch terlama lebih dulu.
-- Poin per transaksi = FLOOR(total_bayar / 10000) x 1,5 (GOLD) atau x 1 (SILVER).
-- Total poin aktif & belum kedaluwarsa = 1.058 poin = Rp1.058.000.
INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status) VALUES
-- Batch warisan (belum ada transaksi pencatatnya)
(1, NULL, 300, 268, DATE_SUB(NOW(), INTERVAL 5 MONTH), DATE_ADD(DATE_SUB(NOW(), INTERVAL 5 MONTH), INTERVAL 6 MONTH), 'AKTIF'),
(2, NULL, 200, 200, DATE_SUB(NOW(), INTERVAL 5 MONTH), DATE_ADD(DATE_SUB(NOW(), INTERVAL 5 MONTH), INTERVAL 6 MONTH), 'AKTIF'),
(3, NULL, 400, 400, DATE_SUB(NOW(), INTERVAL 5 MONTH), DATE_ADD(DATE_SUB(NOW(), INTERVAL 5 MONTH), INTERVAL 6 MONTH), 'AKTIF'),
(4, NULL, 100,  95, DATE_SUB(NOW(), INTERVAL 5 MONTH), DATE_ADD(DATE_SUB(NOW(), INTERVAL 5 MONTH), INTERVAL 6 MONTH), 'AKTIF'),
-- Poin dari masing-masing transaksi
(1, 1,   6,   6, DATE_SUB(NOW(), INTERVAL 118 DAY), DATE_ADD(DATE_SUB(NOW(), INTERVAL 118 DAY), INTERVAL 6 MONTH), 'AKTIF'),
(2, 2,   6,   6, DATE_SUB(NOW(), INTERVAL 110 DAY), DATE_ADD(DATE_SUB(NOW(), INTERVAL 110 DAY), INTERVAL 6 MONTH), 'AKTIF'),
(3, 3,  10,  10, DATE_SUB(NOW(), INTERVAL 96 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 96 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(4, 4,   6,   6, DATE_SUB(NOW(), INTERVAL 84 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 84 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(1, 5,   9,   9, DATE_SUB(NOW(), INTERVAL 72 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 72 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(2, 6,   4,   4, DATE_SUB(NOW(), INTERVAL 62 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 62 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(3, 7,  13,  13, DATE_SUB(NOW(), INTERVAL 50 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 50 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(4, 8,   5,   5, DATE_SUB(NOW(), INTERVAL 40 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 40 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(1, 9,  13,  13, DATE_SUB(NOW(), INTERVAL 30 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 30 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(3, 10,  7,   7, DATE_SUB(NOW(), INTERVAL 20 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 20 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(2, 11,  7,   7, DATE_SUB(NOW(), INTERVAL 12 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 12 DAY),  INTERVAL 6 MONTH), 'AKTIF'),
(1, 12,  9,   9, DATE_SUB(NOW(), INTERVAL 5 DAY),   DATE_ADD(DATE_SUB(NOW(), INTERVAL 5 DAY),   INTERVAL 6 MONTH), 'AKTIF'),
-- Kasus khusus agar laporan terbukti memfilter dengan benar
(2, NULL, 50,  50, DATE_SUB(NOW(), INTERVAL 8 MONTH), DATE_SUB(NOW(), INTERVAL 2 MONTH), 'AKTIF'),
(1, NULL, 10,   0, DATE_SUB(NOW(), INTERVAL 4 MONTH), DATE_ADD(DATE_SUB(NOW(), INTERVAL 4 MONTH), INTERVAL 6 MONTH), 'AKTIF'),
(3, 13,   7,   7, DATE_SUB(NOW(), INTERVAL 2 DAY),  DATE_ADD(DATE_SUB(NOW(), INTERVAL 2 DAY),  INTERVAL 6 MONTH), 'DIBATALKAN');

-- DATA SEED USER SIMULASI (Password: admin123, kasir123, member123)
INSERT INTO users (id, nama, email, password_hash, role) VALUES
(1, 'Pemilik Kedai (Admin)', 'admin@seduhkopi.com', '$2a$10$KPQoj7P.EAmlSLoSAAfFtuMF42IOayaW2jtlLMMGacBwFja/xk6Me', 'ADMIN'),
(2, 'Kasir Kedai Seduh', 'kasir@seduhkopi.com', '$2a$10$sPRcFmGKAwipGlTTJ/62UO7NRWEfB108zfbB34wKqfQ61hDEre6N6', 'KASIR'),
(3, 'Budi (Member Loyal)', 'member@seduhkopi.com', '$2a$10$qIbJFbrmTU.sRWp7JC5RiO6bOvntjkPlUf9Iswa/PtiKKl.RUMox6', 'MEMBER');
