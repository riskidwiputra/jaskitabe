package penyewa

import "gorm.io/gorm"

type Penyewa struct {
	gorm.Model
	Nama    string `json:"nama"`
	NoHp    string `json:"no_hp"`
	Alamat  string `json:"alamat"`
	Catatan string `json:"catatan"`
}
