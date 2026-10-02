package barang

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo *Repository }

func RegisterRoutes(router gin.IRouter, repo *Repository) {
	h := &Handler{repo: repo}
	group := router.Group("/barang")
	group.GET("", h.list)
	group.POST("", h.create)
	group.PUT("/:id", h.update)
	group.DELETE("/:id", h.delete)
}

// prefixForType menentukan huruf depan kode barang berdasarkan tipenya.
// Kalau ada tipe baru yang belum terdaftar di sini, fallback ke "BR"
// supaya tetap jalan (bukan error), tinggal tambah baris baru kalau perlu.
func prefixForType(t string) string {
	switch t {
	case "Jas Pria":
		return "JP"
	case "Celana Pria":
		return "CP"
	case "Jas Wanita":
		return "JW"
	case "Celana Wanita":
		return "CW"
	case "Rok Wanita":
		return "RW"
	case "Dasi":
		return "DS"
	default:
		return "BR"
	}
}

func (h *Handler) list(c *gin.Context) {
	data, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// create sekarang mengabaikan kode_barang yang mungkin ikut terkirim dari
// client, dan selalu men-generate sendiri berdasarkan tipe barangnya.
func (h *Handler) create(c *gin.Context) {
	var input Barang
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kode, err := h.repo.NextKode(prefixForType(input.Type))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat kode barang"})
		return
	}
	input.KodeBarang = kode

	if err := h.repo.Create(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": input})
}

func (h *Handler) update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "barang tidak ditemukan"})
		return
	}
	var input Barang
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = existing.ID
	input.CreatedAt = existing.CreatedAt
	// kode_barang TIDAK diubah lewat edit — tetap pakai yang sudah ada,
	// biar kode barang jadi identitas permanen, tidak berubah-ubah.
	input.KodeBarang = existing.KodeBarang
	if err := h.repo.Update(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": input})
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	deletedBy, _ := c.Get("email")
	if err := h.repo.SoftDelete(uint(id), deletedBy.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "barang dihapus"})
}