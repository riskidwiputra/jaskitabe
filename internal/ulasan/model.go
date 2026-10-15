package ulasan

import (
	"gorm.io/gorm"
	"time"
)

type Token struct {
	gorm.Model
	TransaksiSewaID uint       `gorm:"uniqueIndex;not null" json:"transaksi_sewa_id"`
	TokenHash       string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt       time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt          *time.Time `json:"used_at"`
}

type Review struct {
	gorm.Model
	TransaksiSewaID uint          `gorm:"uniqueIndex;not null" json:"transaksi_sewa_id"`
	NamaPenyewa     string        `gorm:"size:180;not null" json:"nama_penyewa"`
	ItemSewa        string        `gorm:"size:500" json:"item_sewa"`
	TanggalSewa     time.Time     `json:"tanggal_sewa"`
	TanggalKembali  time.Time     `json:"tanggal_kembali"`
	Rating          int           `gorm:"not null" json:"rating"`
	KritikSaran     string        `gorm:"type:text;not null" json:"kritik_saran"`
	Foto            []ReviewPhoto `gorm:"foreignKey:ReviewID" json:"foto"`
}

type ReviewPhoto struct {
	gorm.Model
	ReviewID uint   `gorm:"index;not null" json:"review_id"`
	URL      string `gorm:"size:500;not null" json:"url"`
}
