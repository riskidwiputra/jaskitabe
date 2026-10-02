package penyewa

import "gorm.io/gorm"

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindAll() ([]Penyewa, error) {
	var data []Penyewa
	err := r.db.Find(&data).Error
	return data, err
}

func (r *Repository) Create(p *Penyewa) error { return r.db.Create(p).Error }

func (r *Repository) FindByID(id uint) (*Penyewa, error) {
	var data Penyewa
	err := r.db.First(&data, id).Error
	return &data, err
}
