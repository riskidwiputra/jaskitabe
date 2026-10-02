package barang

import (
	"time"

	"gorm.io/gorm"
)

type Barang struct {
	gorm.Model
	KodeBarang    string    `json:"kode_barang"`
	NamaBarang    string    `json:"nama_barang"`
	Ukuran        string    `json:"ukuran"`
	JumlahBarang  int       `json:"jumlah_barang"`
	HargaSatuan   int       `json:"harga_satuan"`
	Type          string    `json:"type"`
	Status        string    `json:"status" gorm:"default:tersedia"`
	DibeliOleh    string    `json:"dibeli_oleh"`
	TanggalBeli   time.Time `json:"tanggal_beli"`
	Merk          string    `json:"merk"`
	Toko          string    `json:"toko"`
	Keterangan    string    `json:"keterangan"`
	LinkPembelian string    `json:"link_pembelian"`
	DeletedBy     string    `json:"deleted_by"`
}
