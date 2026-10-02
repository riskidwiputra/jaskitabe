package transaksi

import (
	"time"

	"gorm.io/gorm"

	"sewa-jas-api/internal/barang"
	"sewa-jas-api/internal/penyewa"
)

type TransaksiSewa struct {
	gorm.Model
	PenyewaID             *uint            `json:"penyewa_id"`
	Penyewa               *penyewa.Penyewa `json:"penyewa" gorm:"foreignKey:PenyewaID"`
	NamaPenyewaManual     string           `json:"nama_penyewa_manual"`
	NoHpManual            string           `json:"no_hp_manual"`
	JenisJaminan          string           `json:"jenis_jaminan"`
	FotoPeminjam          string           `json:"foto_peminjam" gorm:"type:longtext"`
	Keterangan            string           `json:"keterangan"`
	DeletedBy             string           `json:"deleted_by"`
	TanggalSewa           time.Time        `json:"tanggal_sewa"`
	TanggalKembaliRencana time.Time        `json:"tanggal_kembali_rencana"`
	TanggalKembaliAktual  *time.Time       `json:"tanggal_kembali_aktual"`
	Status                string           `json:"status" gorm:"default:berjalan"`
	HargaSewa             int              `json:"harga_sewa"`
	Barang                []TransaksiSewaBarang `json:"barang" gorm:"foreignKey:TransaksiSewaID"`
}

type TransaksiSewaBarang struct {
	gorm.Model
	TransaksiSewaID uint          `json:"transaksi_sewa_id"`
	BarangID        uint          `json:"barang_id"`
	Barang          barang.Barang `json:"barang" gorm:"foreignKey:BarangID"`
	HargaItem       int           `json:"harga_item"`
}
