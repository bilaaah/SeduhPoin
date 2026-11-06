# SeduhPoin — Sistem Poin Member dan Promo Kedai Kopi Seduh

Sistem informasi manajemen berbasis web untuk **Kedai Kopi Seduh** (3 Cabang: Pusat, Timur, dan Barat) yang mengintegrasikan pengelolaan master data, konfigurasi promosi multi-kondisi, transaksi kasir POS, loyalitas member berjenjang (Silver & Gold), serta laporan analitik biaya promosi dan poin aktif.

Aplikasi dibangun menggunakan bahasa **Go (Golang)** murni dengan arsitektur modular **Clean/Layered Architecture** (`net/http`, `html/template`), database **MySQL (XAMPP)**, serta antarmuka HTML5 & CSS modern bertema kedai kopi.

---

## ☕ Spesifikasi & Aturan Bisnis (Topik 5)

Kedai Kopi Seduh memiliki tiga cabang operasional. Sistem menghubungkan program keanggotaan dan promosi dengan transaksi penjualan di seluruh cabang:

1. **Perolehan Poin Member**:
   - Pelanggan yang terdaftar sebagai member mendapatkan **1 poin untuk setiap kelipatan Rp10.000** dari total belanja setelah potongan promo (pembulatan ke bawah).
   - Member dengan tingkat **Gold** mendapatkan poin **1,5 kali lipat** dari perolehan standar.
   - Poin memiliki masa berlaku **6 bulan** sejak tanggal diperoleh.
2. **Penukaran Poin (Redemption)**:
   - Nilai tukar: **1 poin = Rp1.000**.
   - Maksimal penukaran poin: **Paling banyak 50% dari total belanja** (setelah diskon promo).
   - Urutan konsumsi poin: **FEFO / FIFO** (poin yang paling dekat masa berlakunya/kedaluwarsa dikonsumsi terlebih dahulu).
3. **Tingkat Member (Membership Tiers)**:
   - Terdiri dari dua tingkat: **Silver** dan **Gold**.
   - Member baru memulai dari tingkat **Silver**.
   - Member dengan akumulasi belanja minimal **Rp1.500.000 dalam 3 bulan terakhir** berhak menjadi **Gold** dan menikmati multiplier poin 1,5x.
4. **Jenis & Ketentuan Promo**:
   - **Diskon Persentase Produk (`PERSENTASE_PRODUK`)**: Potongan harga persentase untuk item produk tertentu.
   - **Potongan Minimal Belanja (`POTONGAN_MIN_BELANJA`)**: Potongan nominal rupiah jika transaksi mencapai batas belanja tertentu.
   - **Beli 2 Gratis 1 (`BUY2_GET1`)**: Promo beli 2 gratis 1 untuk kategori produk tertentu.
   - **Voucher Berkode (`VOUCHER`)**: Kupon berkode unik dengan kuota pemakaian terbatas.
   - *Batasan Fleksibel*: Promo dapat dibatasi berdasarkan rentang tanggal aktif, cabang tertentu (`promo_cabang`), dan jam operasional (*happy hour*: `mulai_jam` - `berakhir_jam`).
   - *Aturan Penggabungan Promo*: Promo produk dapat digabung dengan promo lain, tetapi **potongan minimal belanja TIDAK DAPAT digabung dengan voucher**.
5. **Pembatalan Transaksi**:
   - Jika transaksi dibatalkan:
     - Poin yang diperoleh dari transaksi tersebut ditarik kembali (*void*).
     - Poin yang ditukarkan pada transaksi tersebut dikembalikan secara utuh kepada member.
     - Stok produk dipulihkan dan kuota voucher dikembalikan.
6. **Kebutuhan Laporan**:
   - Laporan biaya setiap promo (total nominal rupiah diskon yang dikeluarkan per promo).
   - Jumlah poin yang masih berlaku beserta nilainya dalam rupiah (`sisa_poin * 1.000`).
   - Perkembangan dan distribusi jumlah member per tingkat (Silver vs Gold).

---

## 👥 Pembagian Lingkup Kerja Tim

| Lingkup | Modul & Tanggung Jawab | Penanggung Jawab | Status |
| :--- | :--- | :--- | :--- |
| **Lingkup 1** | **Master Data**: Pengelolaan produk, kategori, 3 cabang, dan data member | **Nabila** | Selesai |
| **Lingkup 2** | **Pengaturan Promo & Voucher**: Promo multi-kondisi, batasan jam/cabang, dan kuota voucher | **Nabila** | Selesai |
| **Auth** | **Autentikasi & Session**: Register, login bcrypt, cookie session HttpOnly, dan middleware | **Nabila** | Selesai |
| **Lingkup 3** | **Transaksi Penjualan**: Perhitungan promo kombinasi, penukaran poin FEFO, dan akumulasi poin | **Rahma** | Siap Dikerjakan |
| **Lingkup 4** | **Pembatalan Transaksi**: Rollback status, penarikan poin reward, refund poin terpakai, restock | **Hilal** | Siap Dikerjakan |
| **Lingkup 5** | **Lifecycle Poin & Tier Member**: Validasi expired 6 bulan & evaluasi belanja 3 bulan (Silver/Gold) | **Dinda** | Siap Dikerjakan |
| **Lingkup 6** | **Laporan Analitik**: Biaya realisasi promo, saldo poin aktif (Rp), dan statistik member | **Luhur** | Siap Dikerjakan |

---

## 🏛️ Arsitektur Sistem

Sistem menggunakan **Layered Architecture** untuk menjamin keteraturan kode dan memudahkan pengerjaan paralel antar anggota tim tanpa konflik:

```text
poinseduh/
├── cmd/
│   └── web/
│       └── main.go                 # HTTP ServeMux, Server lifecycle, dependency injection
├── db/
│   ├── schema.sql                  # DDL skema lengkap MySQL & seed data awal
│   └── auth.sql                    # Skema tabel users, sessions, dan akun demo
├── internal/
│   ├── config/
│   │   └── config.go               # Pembaca konfigurasi environment (.env)
│   ├── handler/
│   │   ├── auth.go                 # Login, register, logout
│   │   ├── admin.go                # Handler cabang, member, promo, voucher (Lingkup 1 & 2)
│   │   ├── catalog.go              # Handler CRUD produk & kategori (Lingkup 1)
│   │   ├── halaman.go              # Handler dashboard ringkasan
│   │   ├── render.go               # Engine HTML Template renderer
│   │   └── errors.go               # HTTP error helper
│   ├── middleware/
│   │   ├── auth.go                 # Middleware RequireLogin & konteks User
│   │   └── logging.go              # Request HTTP logger
│   ├── model/
│   │   ├── admin.go                # Entity: Cabang, Produk, Member, Promo, Voucher
│   │   ├── user.go                 # Entity User & domain error auth
│   │   └── errors.go               # Sentinel errors
│   ├── repository/
│   │   ├── admin.go                # Query SQL Cabang, Member, Promo, Voucher
│   │   ├── catalog.go              # Query SQL Produk & Kategori
│   │   ├── user.go                 # Query SQL Users
│   │   └── session.go              # Query SQL Sessions
│   └── service/
│       └── auth.go                 # Logika bisnis autentikasi bcrypt & session
├── templates/                      # UI Views (HTML5 + CSS Bertema Kopi)
│   ├── auth/
│   │   ├── login.html              # Halaman login modern
│   │   └── register.html           # Halaman registrasi akun
│   ├── layout.html                 # Master layout & responsive stylesheet
│   ├── header.html                 # Topbar header
│   ├── sidebar.html                # Sidebar navigasi profesional
│   ├── footer.html                 # Footer aplikasi
│   ├── home.html                   # Dashboard ringkasan eksekutif
│   ├── produk.html                 # Lingkup 1: UI CRUD Produk & Kategori
│   ├── cabang.html                 # Lingkup 1: UI Cabang Kedai
│   ├── member.html                 # Lingkup 1: UI Registrasi & List Member
│   ├── promo.html                  # Lingkup 2: UI Pengaturan Promo
│   └── voucher.html                # Lingkup 2: UI Voucher Berkode
├── .env                            # Konfigurasi aktif lokal
├── .env.example                    # Template konfigurasi
├── .gitignore                      # File pengecualian Git
├── go.mod                          # Definisi Go module
└── README.md                       # Dokumentasi sistem & panduan teknis tim
```

---

## 🗄️ Skema Database (MySQL / XAMPP)

Tabel-tabel database `seduhpoin` telah dirancang dan dinormalisasi untuk memenuhi seluruh kebutuhan sistem:

```mermaid
erDiagram
    USERS ||--o{ SESSIONS : "memiliki"
    CABANG ||--o{ TRANSAKSI : "mencatat"
    CABANG ||--o{ PROMO_CABANG : "filter cabang"
    MEMBER ||--o{ TRANSAKSI : "berbelanja"
    MEMBER ||--o{ POIN_MEMBER : "memiliki saldo"
    KATEGORI ||--o{ PRODUK : "mengelompokkan"
    KATEGORI ||--o{ PROMO_KATEGORI : "target promo"
    PRODUK ||--o{ PROMO_PRODUK : "target diskon"
    PROMO ||--o{ PROMO_CABANG : "filter cabang"
    PROMO ||--o{ PROMO_PRODUK : "filter produk"
    PROMO ||--o{ PROMO_KATEGORI : "filter kategori"
    PROMO ||--o{ VOUCHER : "diterbitkan sebagai"
    TRANSAKSI ||--o{ TRANSAKSI_ITEM : "berisi item"
    TRANSAKSI_ITEM ||--|| PRODUK : "merujuk produk"
    TRANSAKSI ||--o{ POIN_MEMBER : "menghasilkan / memakai"

    USERS {
        bigint id PK
        varchar nama
        varchar email UK
        varchar password_hash
        enum role "ADMIN, KASIR, MEMBER"
        timestamp created_at
    }

    SESSIONS {
        varchar token PK
        bigint user_id FK
        datetime expires_at
    }

    CABANG {
        bigint id PK
        varchar kode UK
        varchar nama
        varchar alamat
        varchar telepon
        tinyint aktif
    }

    KATEGORI {
        bigint id PK
        varchar nama UK
    }

    PRODUK {
        bigint id PK
        bigint kategori_id FK
        varchar nama
        int harga
        int stok
        text keterangan
        tinyint aktif
        timestamp deleted_at
    }

    MEMBER {
        bigint id PK
        varchar nomor_member UK
        varchar nama
        varchar no_hp UK
        varchar email
        enum tingkat "SILVER, GOLD"
        bigint total_belanja_3_bulan
        tinyint aktif
    }

    PROMO {
        bigint id PK
        varchar kode UK
        varchar nama
        enum jenis "PERSENTASE_PRODUK, POTONGAN_MIN_BELANJA, BUY2_GET1, VOUCHER"
        int nilai
        int minimal_belanja
        int maksimal_pemakaian
        int terpakai
        datetime mulai
        datetime berakhir
        time mulai_jam
        time berakhir_jam
        tinyint aktif
    }

    VOUCHER {
        bigint id PK
        bigint promo_id FK
        varchar kode UK
        int kuota
        int terpakai
    }

    TRANSAKSI {
        bigint id PK
        bigint cabang_id FK
        bigint member_id FK
        bigint voucher_id FK
        int poin_digunakan
        int subtotal
        int total_diskon
        int total_bayar
        enum status "SELESAI, DIBATALKAN"
        timestamp created_at
    }

    TRANSAKSI_ITEM {
        bigint id PK
        bigint transaksi_id FK
        bigint produk_id FK
        int qty
        int harga_satuan
        int diskon_satuan
        int subtotal_item
    }

    POIN_MEMBER {
        bigint id PK
        bigint member_id FK
        bigint transaksi_id FK
        int jumlah
        int sisa
        datetime diperoleh_at
        datetime kadaluarsa_at
        enum status "AKTIF, DIBATALKAN"
    }
```

---

## 📌 Status Implementasi Modul Selesai (Nabila)

### 1. Lingkup 1 — Pengelolaan Data Produk, Cabang, dan Member
- **Menu & Stok Produk (`/produk`)**: CRUD produk lengkap terhubung relasi kategori, penetapan harga jual, manajemen kuantitas stok, dan mekanisme soft-delete (`deleted_at`).
- **Cabang Kedai (`/cabang`)**: Manajemen data 3 cabang resmi (**Seduh Pusat**, **Seduh Timur**, dan **Seduh Barat**) lengkap dengan kode identifikasi, nomor kontak, dan alamat operasional.
- **Data Member (`/member`)**: Pendaftaran kartu anggota loyalitas, format kontak unik, penanda tingkatan **Silver** (standar) dan **Gold** (belanja $\ge$ Rp1.500.000 dalam 3 bulan terakhir).

### 2. Lingkup 2 — Pengaturan Promo dan Voucher
- **Pengaturan Promo Multi-Kondisi (`/promo`)**:
  - 4 jenis promo: Diskon Persentase Produk, Potongan Min. Belanja, Buy 2 Get 1 Kategori, dan Promo Berbasis Voucher.
  - Batasan fleksibel: Rentang tanggal berlaku, jam operasional tertentu (*happy hour*), dan seleksi target cabang maupun item produk/kategori.
- **Kupon Voucher Berkode (`/voucher`)**:
  - Penerbitan kode voucher eksklusif dengan pembatasan kuota klaim maksimal.

### 3. Autentikasi & Session Multi-Peran (Auth Module)
- Login berbasis cookie session `HttpOnly` dengan keamanan `SameSite=Lax`.
- Hashing password menggunakan algoritma `bcrypt`.
- Mendukung 3 peran pengguna:
  - **👑 Admin / Pemilik**: Hak akses penuh kelola master data, promo, voucher, member, dan memantau analitik.
  - **☕ Kasir**: Menangani kasir POS transaksi, klaim promo, penukaran poin, dan pembatalan transaksi.
  - **👤 Member**: Memiliki nomor anggota, cek reward poin, tiering, dan riwayat belanja.

**Kredensial Akun untuk Simulasi:**
- **Admin/Pemilik**: `admin@seduhkopi.com` | Password: `admin123`
- **Kasir**: `kasir@seduhkopi.com` | Password: `kasir123`
- **Member**: `member@seduhkopi.com` | Password: `member123`

---

## 🚀 Panduan Teknis untuk Anggota Tim (Lingkup 3, 4, 5, 6)

Berikut panduan langkah pengerjaan yang terstruktur untuk rekan-rekan tim:

---

### 🔹 Lingkup 3: Transaksi Penjualan (Rahma)
- **Tujuan**: Membangun modul kasir penjualan (POS) di cabang terpilih yang menghitung kalkulasi promo, penukaran poin, dan perolehan poin baru.
- **Berkas yang Perlu Dibuat / Dimodifikasi**:
  - `internal/handler/transaksi.go` (routing `GET /transaksi` & `POST /transaksi`)
  - `internal/repository/transaksi.go`
  - `templates/transaksi.html`
  - Daftarkan route di `cmd/web/main.go` dan tambahkan link di `templates/sidebar.html`.
- **Tabel Database yang Dipakai**: `transaksi`, `transaksi_item`, `produk`, `promo`, `promo_produk`, `promo_kategori`, `promo_cabang`, `voucher`, `poin_member`, `member`.
- **Langkah Logika Bisnis**:
  1. **Form Input**: Pilih cabang, cari member (opsional), pilih produk + qty, kode voucher (opsional), centang tukarkan poin.
  2. **Hitung Subtotal**: Total dari $\sum (\text{harga} \times \text{qty})$.
  3. **Evaluasi Diskon Promo** (cek aktif, dalam rentang tanggal & jam, berlaku untuk cabang ini):
     - *Diskon Persentase Produk (`PERSENTASE_PRODUK`)*: Cek tabel `promo_produk`. Potong `harga_satuan` sebesar `nilai`%. Simpan ke `diskon_satuan` di `transaksi_item`.
     - *Beli 2 Gratis 1 (`BUY2_GET1`)*: Cek `promo_kategori`. Untuk setiap 3 item dalam kategori target, 1 item dihitung gratis. `diskon_satuan = harga_satuan` untuk item gratis.
     - *Potongan Min. Belanja (`POTONGAN_MIN_BELANJA`)*: Jika subtotal ≥ `minimal_belanja` **DAN** tidak ada voucher digunakan.
     - *Voucher (`VOUCHER`)*: Validasi kode, kuota, cabang. Kurangi `voucher.terpakai += 1`. **Aturan mutlak**: jika voucher dipakai, promo `POTONGAN_MIN_BELANJA` tidak boleh aktif.
  4. **Penukaran Poin Member** (jika member dipilih dan meminta tukar poin):
     - Ambil poin valid: `WHERE member_id=? AND status='AKTIF' AND sisa>0 AND kadaluarsa_at > NOW() ORDER BY kadaluarsa_at ASC` (FEFO).
     - **Batas maksimal**: poin yang ditukar ≤ `floor(total_setelah_diskon * 0.5 / 1000)` poin.
     - Kurangi `sisa` per baris poin dari yang paling dekat kedaluwarsa. Catat `poin_digunakan` di tabel `transaksi`.
  5. **Perolehan Poin Baru**:
     - `poin_baru = floor(total_bayar / 10000)`.
     - Jika member `GOLD`: `poin_baru = floor(poin_baru * 1.5)`.
     - `INSERT INTO poin_member (member_id, transaksi_id, jumlah, sisa, diperoleh_at, kadaluarsa_at, status) VALUES (?, ?, ?, ?, NOW(), NOW() + INTERVAL 6 MONTH, 'AKTIF')`.
  6. **Simpan Detail Item**: Insert baris ke `transaksi_item` per produk (untuk kebutuhan restock Hilal).
  7. **Kurangi Stok**: `UPDATE produk SET stok = stok - ? WHERE id = ?`.

---

### 🔹 Lingkup 4: Pembatalan Transaksi (Hilal)
- **Tujuan**: Membatalkan transaksi penjualan jika terjadi kesalahan atau pembatalan pesanan oleh pelanggan.
- **Berkas yang Perlu Dibuat / Dimodifikasi**:
  - `internal/handler/batal.go` (routing `POST /transaksi/{id}/batal`)
  - `internal/repository/batal.go`
  - Tampilkan tombol "Batalkan Transaksi" pada riwayat transaksi.
- **Tabel Database yang Dipakai**: `transaksi`, `transaksi_item`, `produk`, `poin_member`, `voucher`.
- **Langkah Logika Bisnis (Harus dalam 1 Database Transaction / `db.BeginTx`)**:
  1. Validasi transaksi yang akan dibatalkan (pastikan status saat ini adalah `SELESAI`).
  2. Ubah status transaksi menjadi `DIBATALKAN`:
     `UPDATE transaksi SET status = 'DIBATALKAN' WHERE id = ?`.
  3. **Tarik Kembali Poin Diperoleh**:
     - Poin yang diterbitkan dari transaksi tersebut di-void:
       `UPDATE poin_member SET status = 'DIBATALKAN', sisa = 0 WHERE transaksi_id = ? AND status = 'AKTIF'`.
  4. **Kembalikan Poin yang Ditukarkan**:
     - Ambil nilai `poin_digunakan` dari tabel `transaksi`. Jika > 0, tambahkan kembali ke saldo poin member (insert baris baru `poin_member` tanpa `transaksi_id` atau pulihkan baris lama).
  5. **Restock Produk**:
     - Query detail item dari `transaksi_item` lalu:
       `UPDATE produk SET stok = stok + ? WHERE id = ?` per baris item.
  6. **Kembalikan Kuota Voucher** (jika transaksi menggunakan voucher):
     - `UPDATE voucher SET terpakai = terpakai - 1 WHERE id = (SELECT voucher_id FROM transaksi WHERE id = ?)`.

---

### 🔹 Lingkup 5: Pengelolaan Masa Berlaku Poin & Tingkat Member (Dinda)
- **Tujuan**: Menangani validasi kedaluwarsa poin (6 bulan) dan evaluasi otomatis tingkat keanggotaan (Silver vs Gold).
- **Berkas yang Perlu Dibuat / Dimodifikasi**:
  - `internal/handler/member_tier.go` (routing `GET /member/tier`)
  - `internal/repository/member_tier.go`
  - `templates/poin_tier.html` (halaman monitoring saldo poin per batch & status tier member).
- **Langkah Logika Bisnis**:
  1. **Masa Berlaku Poin (6 Bulan)**:
     - Tampilkan riwayat batch poin per member: tanggal perolehan, tanggal kedaluwarsa, jumlah awal, sisa poin, dan status (Aktif/Kedaluwarsa/Dibatalkan).
     - Query poin aktif valid:
       ```sql
       SELECT member_id, SUM(sisa) AS poin_aktif, (SUM(sisa) * 1000) AS nilai_rupiah
       FROM poin_member
       WHERE status = 'AKTIF' AND kadaluarsa_at > NOW()
       GROUP BY member_id;
       ```
  2. **Evaluasi Tingkat Keanggotaan (Silver & Gold)**:
     - Hitung akumulasi belanja riil member dalam 3 bulan terakhir:
       ```sql
       SELECT COALESCE(SUM(total_bayar), 0)
       FROM transaksi
       WHERE member_id = ? AND status = 'SELESAI'
         AND created_at >= NOW() - INTERVAL 3 MONTH;
       ```
     - **Aturan Evaluasi:**
       - Jika total belanja 3 bulan $\ge$ **Rp1.500.000** $\rightarrow$ Status diubah/dipertahankan menjadi **GOLD**.
       - Jika total belanja 3 bulan $<$ **Rp1.500.000** $\rightarrow$ Status menjadi **SILVER**.
     - Simpan hasil perhitungan ke `UPDATE member SET tingkat = ?, total_belanja_3_bulan = ? WHERE id = ?`.

---

### 🔹 Lingkup 6: Laporan Promo, Poin, dan Member (Luhur)
- **Tujuan**: Membangun modul pelaporan analitik bisnis untuk pemilik kedai.
- **Berkas yang Perlu Dibuat / Dimodifikasi**:
  - `internal/handler/laporan.go` (routing `GET /laporan`)
  - `internal/repository/laporan.go`
  - `templates/laporan.html`
- **3 Laporan Wajib Sesuai Soal Topik 5**:
  1. **Laporan Biaya Setiap Promo**:
     - Menghitung total realisasi rupiah potongan diskon yang diserap oleh masing-masing promo.
     - Rekap berdasarkan kode promo, jenis promo, jumlah pemakaian, dan total nominal rupiah diskon yang dikeluarkan kedai.
  2. **Laporan Poin yang Masih Berlaku & Nilai Rupiahnya**:
     - Menampilkan total poin aktif yang masih berlaku di seluruh cabang:
       ```sql
       SELECT 
           COUNT(DISTINCT member_id) AS total_member_berpoin,
           COALESCE(SUM(sisa), 0) AS total_poin_aktif,
           COALESCE(SUM(sisa) * 1000, 0) AS total_nilai_rupiah
       FROM poin_member
       WHERE status = 'AKTIF' AND kadaluarsa_at > NOW();
       ```
  3. **Laporan Perkembangan Jumlah Member per Tingkat**:
     - Rekapitulasi perbandingan total member **Silver** vs **Gold**.
     - Grafik atau tabel distribusi keanggotaan dan pertumbuhan member baru per bulan.

---

## 🛠️ Panduan Menjalankan Aplikasi di Lokal (XAMPP)

1. **Jalankan Apache & MySQL di XAMPP Control Panel**.
2. **Import Database (Cukup 1 Berkas)**:
   - Buka phpMyAdmin di `http://localhost/phpmyadmin`.
   - Import berkas [`db/schema.sql`](db/schema.sql). Database `seduhpoin`, seluruh tabel, dan data simulasi awal akan otomatis terbuat.
3. **Konfigurasi `.env`**:
   Pastikan berkas `.env` ada di root direktori dengan isi:
   ```env
   DB_DSN=root:@tcp(127.0.0.1:3306)/seduhpoin?parseTime=true&charset=utf8mb4
   ADDR=:8080
   COOKIE_SECURE=false
   ```
4. **Jalankan Web Server**:
   ```bash
   go run ./cmd/web
   ```
5. **Akses Aplikasi**:
   Buka peramban di `http://localhost:8080`.
   Gunakan akun demo:
   - Email: `admin@seduhkopi.com`
   - Password: `admin123`

---

## 🌿 Panduan Kolaborasi Git Tim

> **Alur wajib**: semua anggota push ke branch `dev` terlebih dahulu. Branch `main` hanya diupdate oleh Nabila (ketua) setelah semua lingkup di `dev` sudah berjalan & direview.

```
origin/main   ← versi stabil final (hanya Nabila yang merge ke sini)
      ↑
origin/dev    ← branch integrasi tim (semua push & pull request ke sini)
      ↑
feat/lingkup-3-transaksi  ← branch kerja masing-masing anggota
feat/lingkup-4-batal
feat/lingkup-5-poin-tier
feat/lingkup-6-laporan
```

1. **Clone Repo & Pindah ke `dev`**:
   ```bash
   git clone https://github.com/bilaaah/SeduhPoin.git
   cd SeduhPoin
   git checkout dev
   ```
2. **Buat Branch Kerja Khusus dari `dev`**:
   ```bash
   git checkout -b fitur/lingkup-3-transaksi    # Rahma
   git checkout -b fitur/lingkup-4-batal        # Hilal
   git checkout -b fitur/lingkup-5-poin-tier    # Dinda
   git checkout -b fitur/lingkup-6-laporan      # Luhur
   ```
3. **Commit dengan Pesan Deskriptif**:
   ```bash
   git add .
   git commit -m "feat(lingkup-3): kalkulasi diskon promo dan penukaran poin FEFO"
   ```
4. **Sinkronisasi dengan `dev` Sebelum Push** (hindari konflik):
   ```bash
   git pull origin dev --rebase
   git push origin fitur/lingkup-3-transaksi
   ```
5. **Buat Pull Request ke `dev`** (bukan ke `main`) melalui GitHub, lalu minta Nabila untuk mereview.
