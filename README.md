# Sewa Jas API — Backend lengkap (rebuild)

## Setup
1. Buat database MySQL `sewa_jas`
2. `cp .env.example .env` lalu sesuaikan kredensial
3. `go mod tidy`
4. `go run cmd/api/main.go`

## Fitur yang ada
- Auth (register/login, JWT)
- Barang: CRUD + soft delete (kode, ukuran, jumlah, harga, tipe, tanggal beli, merk, toko, link pembelian, dibeli oleh)
- Penyewa: create + list
- Transaksi Sewa: CRUD (penyewa opsional/manual, jaminan, foto base64, banyak barang), tandai kembali (otomatis update status barang)
- Denda: CRUD, terhubung ke transaksi
- Pengeluaran: CRUD, kategori "Modal" vs operasional, dikeluarkan oleh
- Notifikasi jatuh tempo
- Soft delete + jejak siapa yang menghapus (`deleted_by`) di semua fitur CRUD
- Ulasan pelanggan: link ulasan untuk transaksi selesai, galeri publik, unggah maksimal 5 foto per ulasan ke Cloudflare R2, pengaktifan kembali link, dan pengelolaan ulasan admin

Untuk link ulasan, set `FRONTEND_URL` ke alamat frontend (misalnya `http://localhost:3000`). Sebelum upload foto, buat bucket R2, aktifkan akses publik melalui custom domain atau URL publik `r2.dev`, lalu buat token S3 dengan izin **Object Read & Write** yang dibatasi ke bucket tersebut. `R2_PUBLIC_URL` adalah URL publik dasar yang membuka objek, bukan endpoint S3. Atur variabel berikut di environment backend; jangan commit access key/secret:

```env
R2_ACCOUNT_ID=account-id-cloudflare
R2_ACCESS_KEY_ID=access-key
R2_SECRET_ACCESS_KEY=secret-key
R2_BUCKET_NAME=jaskita-ulasan
R2_PUBLIC_URL=https://images.example.com
```

`R2_PUBLIC_URL` harus berupa base URL publik untuk objek (tanpa garis miring di akhir). Foto lama yang tersimpan di `uploads/ulasan` tetap dilayani dari backend dan dapat dihapus oleh admin; foto baru diunggah ke R2. Upload akan memberi pesan konfigurasi yang jelas jika environment R2 belum lengkap.

## Deployment gratis (untuk uji coba)

Konfigurasi `render.yaml` menyiapkan API di Render Free. Backend tetap memakai driver MySQL; database TiDB Cloud Starter yang dipilih untuk panduan ini kompatibel dengan protokol MySQL, bukan PostgreSQL atau server MySQL yang dikelola sendiri.

1. Buat instance **MySQL-compatible TiDB Cloud Starter**. Simpan host, username, password, dan nama database dari halaman koneksi. Pastikan endpoint publik mengizinkan koneksi dari API Render; koneksi dari aplikasi menggunakan TLS.
2. Hubungkan repo backend ini ke Render sebagai Blueprint. Isi `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, dan `JWT_SECRET` saat diminta. `DB_PORT=4000` dan `DB_TLS=true` sudah disiapkan. Buat `JWT_SECRET` acak yang kuat; jangan masukkan rahasia ke Git.
3. Setelah API berhasil deploy, buka `https://<nama-service>.onrender.com/health`; hasil sehat adalah `{"status":"ok"}`.
4. Deploy repo frontend di Netlify, lalu atur environment variable `NEXT_PUBLIC_API_URL` ke `https://<nama-service>.onrender.com/api`. Setelah URL frontend diketahui, isi `FRONTEND_URL` pada environment Render dengan URL tersebut lalu deploy ulang.
5. Agar unggah foto ulasan berfungsi dan tetap tersimpan, isi `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, dan `R2_PUBLIC_URL` di environment Render. Jangan commit access key atau secret.

Layanan gratis memiliki batas penggunaan dan tidak ditujukan untuk produksi dengan kebutuhan ketersediaan tinggi. Render Free menidurkan API setelah tidak aktif, sehingga permintaan pertama setelah jeda dapat lambat. File lokal di `uploads/ulasan` bersifat sementara di Render; gunakan R2 untuk foto yang perlu bertahan. TiDB Cloud Starter menyediakan kuota bulanan gratis; jika batas kuota tercapai, koneksi/operasi dapat dibatasi. Periksa batas dan kebijakan harga terbaru sebelum mengandalkannya untuk operasional usaha.
# jaskitabe
