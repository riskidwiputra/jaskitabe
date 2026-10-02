package pengeluaran

import (
	"time"

	"gorm.io/gorm"
)

type Pengeluaran struct {
	gorm.Model
	Tanggal         time.Time `json:"tanggal"`
	Kategori        string    `json:"kategori"`
	Jumlah          int       `json:"jumlah"`
	Keterangan      string    `json:"keterangan"`
	DikeluarkanOleh string    `json:"dikeluarkan_oleh"`
	DeletedBy       string    `json:"deleted_by"`
}
