-- ============================================================================
-- MIGRASI LINGKUP 6 - LAPORAN ANALITIK (Luhur)
-- ----------------------------------------------------------------------------
-- File ini untuk instalasi yang database-nya SUDAH ADA dan tidak ingin di-drop.
-- db/schema.sql tetap menjadi sumber utama (import ulang = data demo lengkap
-- dengan transaksi simulasi). Jalankan file ini hanya bila tabel
-- transaksi_promo belum ada di database yang sedang dipakai.
--
-- Cara pakai: buka phpMyAdmin > database seduhpoin > tab SQL > import file ini.
-- Script ini idempotent, boleh dijalankan berkali-kali tanpa efek ganda.
-- ============================================================================

USE seduhpoin;

-- 1. Ledger pemakaian promo: sumber kebenaran "biaya realisasi promo".
--    Satu baris = satu promo yang benar-benar dipakai pada satu transaksi,
--    beserta nominal rupiah diskon yang diserap.
CREATE TABLE IF NOT EXISTS transaksi_promo (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    transaksi_id BIGINT UNSIGNED NOT NULL,
    promo_id BIGINT UNSIGNED NOT NULL,
    jenis ENUM('PERSENTASE_PRODUK','POTONGAN_MIN_BELANJA','BUY2_GET1','VOUCHER') NOT NULL,
    diskon INT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_transaksi_promo (transaksi_id, promo_id),
    KEY idx_trxpromo_promo (promo_id),
    CONSTRAINT fk_trxpromo_transaksi FOREIGN KEY (transaksi_id) REFERENCES transaksi(id) ON DELETE CASCADE,
    CONSTRAINT fk_trxpromo_promo FOREIGN KEY (promo_id) REFERENCES promo(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. Index untuk query pertumbuhan member per bulan pada Laporan Analitik.
--    Dijaga agar tidak gagal bila index sudah ada.
SET @idx_member_created = (
    SELECT COUNT(*) FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = 'seduhpoin'
      AND TABLE_NAME = 'member'
      AND INDEX_NAME = 'idx_member_created'
);
SET @sql_member_created = IF(
    @idx_member_created = 0,
    'ALTER TABLE member ADD KEY idx_member_created (created_at)',
    'SELECT ''idx_member_created sudah ada'''
);
PREPARE stmt_member_created FROM @sql_member_created;
EXECUTE stmt_member_created;
DEALLOCATE PREPARE stmt_member_created;

-- 3. Catatan: data simulasi (transaksi, transaksi_item, transaksi_promo,
--    poin_member) sengaja TIDAK disertakan di file ini karena memakai id
--    eksplisit yang berpotensi bentrok dengan data produksi. Untuk data demo,
--    import ulang db/schema.sql (database akan dibuat ulang dari nol).
--
-- 4. Verifikasi hasil migrasi:
--    SELECT COUNT(*) FROM transaksi_promo;
--    SHOW INDEX FROM member WHERE Key_name = 'idx_member_created';
