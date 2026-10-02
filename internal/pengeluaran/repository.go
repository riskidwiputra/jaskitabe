package pengeluaran

import "gorm.io/gorm"

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindAll() ([]Pengeluaran, error) {
	var data []Pengeluaran
	err := r.db.Order("tanggal desc").Find(&data).Error
	return data, err
}

func (r *Repository) FindByID(id uint) (*Pengeluaran, error) {
	var p Pengeluaran
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *Repository) Create(p *Pengeluaran) error { return r.db.Create(p).Error }
func (r *Repository) Update(p *Pengeluaran) error { return r.db.Save(p).Error }

func (r *Repository) SoftDelete(id uint, deletedBy string) error {
	if err := r.db.Model(&Pengeluaran{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return r.db.Delete(&Pengeluaran{}, id).Error
}
