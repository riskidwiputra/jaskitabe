package ulasan

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sewa-jas-api/internal/transaksi"
)

var ErrReviewNotFound = errors.New("ulasan tidak ditemukan")

type Repository struct{ db *gorm.DB }

type AdminReview struct {
	TransaksiSewaID uint       `json:"transaksi_sewa_id"`
	StatusLink      string     `json:"status_link"`
	ExpiresAt       time.Time  `json:"expires_at"`
	UsedAt          *time.Time `json:"used_at"`
	Review          *Review    `json:"review"`
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateToken(transaksiID uint, raw, hash string, expires time.Time) error {
	var t transaksi.TransaksiSewa
	if err := r.db.First(&t, transaksiID).Error; err != nil {
		return errors.New("transaksi tidak ditemukan")
	}
	if t.Status != "selesai" {
		return errors.New("link ulasan hanya bisa dibuat untuk transaksi selesai")
	}
	var existing Review
	if err := r.db.Where("transaksi_sewa_id = ?", transaksiID).First(&existing).Error; err == nil {
		return errors.New("transaksi ini sudah memiliki ulasan; aktifkan kembali link dari halaman kelola ulasan")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var token Token
	err := r.db.Where("transaksi_sewa_id = ?", transaksiID).First(&token).Error
	if err == nil && token.UsedAt == nil && token.ExpiresAt.After(time.Now()) {
		return errors.New("link ulasan aktif sudah pernah dibuat untuk transaksi ini")
	}
	if err == nil {
		return r.db.Model(&token).Updates(map[string]interface{}{
			"token_hash": hash,
			"expires_at": expires,
			"used_at":    nil,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(&Token{TransaksiSewaID: transaksiID, TokenHash: hash, ExpiresAt: expires}).Error
}

func (r *Repository) ReactivateToken(transaksiID uint) error {
	var token Token
	if err := r.db.Where("transaksi_sewa_id = ?", transaksiID).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("link ulasan belum pernah dibuat")
		}
		return err
	}
	return r.db.Model(&token).Updates(map[string]interface{}{
		"expires_at": time.Now().Add(30 * 24 * time.Hour),
		"used_at":    nil,
	}).Error
}

func (r *Repository) TransactionInfo(rawHash string) (map[string]interface{}, error) {
	var token Token
	if err := r.db.Where("token_hash = ?", rawHash).First(&token).Error; err != nil {
		return nil, errors.New("link ulasan tidak valid")
	}
	if token.UsedAt != nil {
		return nil, errors.New("link ulasan sudah digunakan")
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("link ulasan sudah kedaluwarsa")
	}
	var t transaksi.TransaksiSewa
	if err := r.db.Preload("Penyewa").Preload("Barang.Barang").First(&t, token.TransaksiSewaID).Error; err != nil {
		return nil, errors.New("transaksi tidak ditemukan")
	}
	if t.Status != "selesai" {
		return nil, errors.New("transaksi belum selesai")
	}
	name := t.NamaPenyewaManual
	if t.Penyewa != nil && t.Penyewa.Nama != "" {
		name = t.Penyewa.Nama
	}
	items := ""
	for i, x := range t.Barang {
		if i > 0 {
			items += ", "
		}
		items += x.Barang.NamaBarang
	}
	var review Review
	var existingReview *Review
	if err := r.db.Preload("Foto").Where("transaksi_sewa_id = ?", t.ID).First(&review).Error; err == nil {
		existingReview = &review
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return map[string]interface{}{
		"nama_penyewa":    name,
		"item_sewa":       items,
		"tanggal_sewa":    t.TanggalSewa,
		"tanggal_kembali": t.TanggalKembaliRencana,
		"expires_at":      token.ExpiresAt,
		"review":          existingReview,
	}, nil
}

func (r *Repository) Submit(hash string, rating int, text string, photoURLs []string) (*Review, []string, error) {
	var result Review
	var oldPhotoURLs []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var token Token
		if err := tx.Where("token_hash = ?", hash).First(&token).Error; err != nil {
			return errors.New("link ulasan tidak valid")
		}
		if token.UsedAt != nil {
			return errors.New("link ulasan sudah digunakan")
		}
		now := time.Now()
		if now.After(token.ExpiresAt) {
			return errors.New("link ulasan sudah kedaluwarsa")
		}
		claimed := tx.Model(&Token{}).
			Where("id = ? AND used_at IS NULL AND expires_at > ?", token.ID, now).
			Update("used_at", now)
		if claimed.Error != nil {
			return claimed.Error
		}
		if claimed.RowsAffected != 1 {
			return errors.New("link ulasan sudah digunakan atau kedaluwarsa")
		}

		var t transaksi.TransaksiSewa
		if err := tx.Preload("Penyewa").Preload("Barang.Barang").First(&t, token.TransaksiSewaID).Error; err != nil {
			return err
		}
		if t.Status != "selesai" {
			return errors.New("transaksi belum selesai")
		}
		name := t.NamaPenyewaManual
		if t.Penyewa != nil && t.Penyewa.Nama != "" {
			name = t.Penyewa.Nama
		}
		items := ""
		for i, x := range t.Barang {
			if i > 0 {
				items += ", "
			}
			items += x.Barang.NamaBarang
		}

		var existing Review
		err := tx.Where("transaksi_sewa_id = ?", t.ID).First(&existing).Error
		if err == nil {
			if len(photoURLs) > 0 {
				var oldPhotos []ReviewPhoto
				if err := tx.Where("review_id = ?", existing.ID).Find(&oldPhotos).Error; err != nil {
					return err
				}
				for _, photo := range oldPhotos {
					oldPhotoURLs = append(oldPhotoURLs, photo.URL)
				}
				if err := tx.Unscoped().Where("review_id = ?", existing.ID).Delete(&ReviewPhoto{}).Error; err != nil {
					return err
				}
			}
			result = existing
			result.NamaPenyewa = name
			result.ItemSewa = items
			result.TanggalSewa = t.TanggalSewa
			result.TanggalKembali = t.TanggalKembaliRencana
			result.Rating = rating
			result.KritikSaran = text
			if err := tx.Save(&result).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			result = Review{
				TransaksiSewaID: t.ID,
				NamaPenyewa:     name,
				ItemSewa:        items,
				TanggalSewa:     t.TanggalSewa,
				TanggalKembali:  t.TanggalKembaliRencana,
				Rating:          rating,
				KritikSaran:     text,
			}
			if err := tx.Create(&result).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		for _, url := range photoURLs {
			if err := tx.Create(&ReviewPhoto{ReviewID: result.ID, URL: url}).Error; err != nil {
				return err
			}
		}
		return tx.Preload("Foto").First(&result, result.ID).Error
	})
	if err != nil {
		return nil, nil, err
	}
	return &result, oldPhotoURLs, nil
}

func (r *Repository) ListForAdmin() ([]AdminReview, error) {
	var tokens []Token
	if err := r.db.Order("created_at desc").Find(&tokens).Error; err != nil {
		return nil, err
	}
	var reviews []Review
	if err := r.db.Preload("Foto").Find(&reviews).Error; err != nil {
		return nil, err
	}
	reviewsByTransaction := make(map[uint]*Review, len(reviews))
	for i := range reviews {
		reviewsByTransaction[reviews[i].TransaksiSewaID] = &reviews[i]
	}

	now := time.Now()
	result := make([]AdminReview, 0, len(tokens))
	for _, token := range tokens {
		status := "belum_digunakan"
		if token.UsedAt != nil {
			status = "sudah_digunakan"
		} else if now.After(token.ExpiresAt) {
			status = "kedaluwarsa"
		}
		result = append(result, AdminReview{
			TransaksiSewaID: token.TransaksiSewaID,
			StatusLink:      status,
			ExpiresAt:       token.ExpiresAt,
			UsedAt:          token.UsedAt,
			Review:          reviewsByTransaction[token.TransaksiSewaID],
		})
	}
	return result, nil
}

func (r *Repository) UpdateReview(id uint, rating int, text string) (*Review, error) {
	var review Review
	if err := r.db.First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	if err := r.db.Model(&review).Updates(map[string]interface{}{
		"rating":       rating,
		"kritik_saran": text,
	}).Error; err != nil {
		return nil, err
	}
	if err := r.db.Preload("Foto").First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *Repository) DeleteReview(id uint) ([]string, error) {
	var photoURLs []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var review Review
		if err := tx.First(&review, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrReviewNotFound
			}
			return err
		}
		var photos []ReviewPhoto
		if err := tx.Where("review_id = ?", id).Find(&photos).Error; err != nil {
			return err
		}
		for _, photo := range photos {
			photoURLs = append(photoURLs, photo.URL)
		}
		if err := tx.Unscoped().Where("review_id = ?", id).Delete(&ReviewPhoto{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&review).Error
	})
	if err != nil {
		return nil, err
	}
	return photoURLs, nil
}

func (r *Repository) List() ([]Review, error) {
	var data []Review
	err := r.db.Preload("Foto").Order("created_at desc").Find(&data).Error
	return data, err
}

func (r *Repository) ListPublic(page, pageSize, rating int) ([]Review, int64, int64, float64, error) {
	var summary struct {
		Total   int64
		Average float64
	}
	if err := r.db.Model(&Review{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating), 0) AS average").
		Scan(&summary).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	query := r.db.Model(&Review{})
	if rating > 0 {
		query = query.Where("rating = ?", rating)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	data := make([]Review, 0, pageSize)
	offset := (page - 1) * pageSize
	if err := query.Preload("Foto").
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&data).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	return data, total, summary.Total, summary.Average, nil
}
