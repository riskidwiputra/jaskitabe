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
# jaskitabe
