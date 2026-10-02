package denda

import "gorm.io/gorm"

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindAll() ([]Denda, error) {
	var data []Denda
	err := r.db.Preload("TransaksiSewa.Penyewa").Find(&data).Error
	return data, err
}

func (r *Repository) FindByTransaksi(transaksiID uint) ([]Denda, error) {
	var data []Denda
	err := r.db.Where("transaksi_sewa_id = ?", transaksiID).Find(&data).Error
	return data, err
}

func (r *Repository) Create(d *Denda) error { return r.db.Create(d).Error }

func (r *Repository) FindByID(id uint) (*Denda, error) {
	var d Denda
	err := r.db.Preload("TransaksiSewa.Penyewa").First(&d, id).Error
	return &d, err
}

func (r *Repository) Update(d *Denda) error { return r.db.Save(d).Error }

func (r *Repository) SoftDelete(id uint, deletedBy string) error {
	if err := r.db.Model(&Denda{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return r.db.Delete(&Denda{}, id).Error
}
