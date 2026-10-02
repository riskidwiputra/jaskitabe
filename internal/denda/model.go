package denda

import (
	"gorm.io/gorm"

	"sewa-jas-api/internal/transaksi"
)

type Denda struct {
	gorm.Model
	TransaksiSewaID uint                    `json:"transaksi_sewa_id"`
	TransaksiSewa   transaksi.TransaksiSewa `json:"transaksi_sewa" gorm:"foreignKey:TransaksiSewaID"`
	Jumlah          int                     `json:"jumlah"`
	Alasan          string                  `json:"alasan"`
	DeletedBy       string                  `json:"deleted_by"`
}
