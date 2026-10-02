package transaksi

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sewa-jas-api/internal/barang"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindAll() ([]TransaksiSewa, error) {
	var data []TransaksiSewa
	err := r.db.Preload("Penyewa").Preload("Barang.Barang").Find(&data).Error
	return data, err
}

func (r *Repository) FindByID(id uint) (*TransaksiSewa, error) {
	var t TransaksiSewa
	err := r.db.Preload("Penyewa").Preload("Barang.Barang").First(&t, id).Error
	return &t, err
}

func (r *Repository) Create(t *TransaksiSewa) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if t.TanggalKembaliRencana.Before(t.TanggalSewa) {
			return errors.New("tanggal kembali tidak boleh lebih awal dari tanggal sewa")
		}
		if t.PenyewaID == nil && t.NamaPenyewaManual == "" {
			return errors.New("pilih penyewa dari data, atau isi nama penyewa manual")
		}
		for _, item := range t.Barang {
			var b barang.Barang
			if err := tx.First(&b, item.BarangID).Error; err != nil {
				return errors.New("barang dengan id tersebut tidak ditemukan")
			}
			if b.Status != "tersedia" {
				return errors.New(b.NamaBarang + " sedang tidak tersedia")
			}
		}
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		for _, item := range t.Barang {
			err := tx.Model(&barang.Barang{}).Where("id = ?", item.BarangID).Update("status", "disewa").Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) Kembalikan(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var t TransaksiSewa
		if err := tx.Preload("Barang").First(&t, id).Error; err != nil {
			return errors.New("transaksi tidak ditemukan")
		}
		now := time.Now()
		t.TanggalKembaliAktual = &now
		t.Status = "selesai"
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
		for _, item := range t.Barang {
			err := tx.Model(&barang.Barang{}).Where("id = ?", item.BarangID).Update("status", "tersedia").Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) UpdateMeta(t *TransaksiSewa) error {
	return r.db.Model(&TransaksiSewa{}).Where("id = ?", t.ID).Updates(map[string]interface{}{
		"nama_penyewa_manual":     t.NamaPenyewaManual,
		"no_hp_manual":            t.NoHpManual,
		"tanggal_kembali_rencana": t.TanggalKembaliRencana,
		"harga_sewa":              t.HargaSewa,
		"jenis_jaminan":           t.JenisJaminan,
		"foto_peminjam":           t.FotoPeminjam,
		"keterangan":              t.Keterangan,
	}).Error
}

func (r *Repository) SoftDelete(id uint, deletedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var t TransaksiSewa
		if err := tx.Preload("Barang").First(&t, id).Error; err != nil {
			return errors.New("transaksi tidak ditemukan")
		}
		if t.Status != "selesai" {
			for _, item := range t.Barang {
				err := tx.Model(&barang.Barang{}).Where("id = ?", item.BarangID).Update("status", "tersedia").Error
				if err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&TransaksiSewa{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error; err != nil {
			return err
		}
		return tx.Delete(&TransaksiSewa{}, id).Error
	})
}

func (r *Repository) FindJatuhTempo() ([]TransaksiSewa, error) {
	var data []TransaksiSewa
	akhirHariIni := time.Now().Format("2006-01-02") + " 23:59:59"
	err := r.db.Preload("Penyewa").Preload("Barang.Barang").
		Where("status = ? AND tanggal_kembali_rencana <= ?", "berjalan", akhirHariIni).
		Find(&data).Error
	return data, err
}
