package barang

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindAll() ([]Barang, error) {
	var data []Barang
	err := r.db.Find(&data).Error
	return data, err
}

func (r *Repository) FindByID(id uint) (*Barang, error) {
	var b Barang
	err := r.db.First(&b, id).Error
	return &b, err
}

func (r *Repository) Create(b *Barang) error { return r.db.Create(b).Error }
func (r *Repository) Update(b *Barang) error { return r.db.Save(b).Error }

func (r *Repository) SoftDelete(id uint, deletedBy string) error {
	if err := r.db.Model(&Barang{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return r.db.Delete(&Barang{}, id).Error
}

// NextKode mencari nomor urut berikutnya untuk sebuah prefix (misal "JP"),
// dengan melihat SEMUA barang yang pernah punya kode berawalan itu —
// termasuk yang sudah dihapus (Unscoped) — supaya nomor tidak pernah
// "dipakai ulang" walau ada barang yang dihapus di tengah jalan.
func (r *Repository) NextKode(prefix string) (string, error) {
	var kodes []string
	err := r.db.Unscoped().Model(&Barang{}).
		Where("kode_barang LIKE ?", prefix+"%").
		Pluck("kode_barang", &kodes).Error
	if err != nil {
		return "", err
	}

	max := 0
	for _, k := range kodes {
		suffix := strings.TrimPrefix(k, prefix)
		if n, convErr := strconv.Atoi(suffix); convErr == nil && n > max {
			max = n
		}
	}

	return fmt.Sprintf("%s%03d", prefix, max+1), nil
}