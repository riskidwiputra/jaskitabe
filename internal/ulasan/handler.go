package ulasan

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo *Repository }

func RegisterRoutes(public, protected gin.IRouter, repo *Repository) {
	h := &Handler{repo: repo}
	admin := protected.Group("/ulasan")
	admin.Use(requireAdmin)
	admin.POST("/transaksi/:id/link", h.createLink)
	admin.PUT("/transaksi/:id/link/aktifkan", h.reactivateLink)
	admin.GET("", h.list)
	admin.PUT("/:id", h.update)
	admin.DELETE("/:id", h.delete)
	public.GET("/ulasan/:token", h.info)
	public.POST("/ulasan/:token", h.submit)
	public.GET("/ulasan", h.publicList)
}

func requireAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "owner" && role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "akses hanya untuk admin"})
		return
	}
	c.Next()
}

func digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

func (h *Handler) createLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "id transaksi tidak valid"})
		return
	}
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		c.JSON(500, gin.H{"error": "gagal membuat token"})
		return
	}
	raw := hex.EncodeToString(b)
	if err = h.repo.CreateToken(uint(id), raw, digest(raw), time.Now().Add(30*24*time.Hour)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	base := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	c.JSON(201, gin.H{"url": base + "/ulasan/" + raw, "expires_at": time.Now().Add(30 * 24 * time.Hour)})
}

func (h *Handler) reactivateLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id transaksi tidak valid"})
		return
	}
	if err := h.repo.ReactivateToken(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "link ulasan diaktifkan kembali"})
}

func (h *Handler) info(c *gin.Context) {
	data, err := h.repo.TransactionInfo(digest(c.Param("token")))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *Handler) submit(c *gin.Context) {
	rating, e := strconv.Atoi(c.PostForm("rating"))
	text := strings.TrimSpace(c.PostForm("kritik_saran"))
	if e != nil || rating < 1 || rating > 5 {
		c.JSON(400, gin.H{"error": "rating wajib dipilih dari 1 sampai 5"})
		return
	}
	if text == "" {
		c.JSON(400, gin.H{"error": "kritik dan saran wajib diisi"})
		return
	}
	if len([]rune(text)) > 5000 {
		c.JSON(400, gin.H{"error": "kritik dan saran maksimal 5000 karakter"})
		return
	}
	files := c.Request.MultipartForm
	urls := []string{}
	if files != nil {
		photos := files.File["foto"]
		if len(photos) > 5 {
			c.JSON(400, gin.H{"error": "maksimal 5 foto"})
			return
		}
		contentTypes := make([]string, len(photos))
		for i, f := range photos {
			if f.Size > 5*1024*1024 {
				c.JSON(400, gin.H{"error": "ukuran tiap foto maksimal 5 MB"})
				return
			}
			ext := strings.ToLower(filepath.Ext(f.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				c.JSON(400, gin.H{"error": "foto harus JPG atau PNG"})
				return
			}
			src, err := f.Open()
			if err != nil {
				c.JSON(400, gin.H{"error": "foto tidak dapat dibaca"})
				return
			}
			head := make([]byte, 512)
			n, readErr := io.ReadFull(src, head)
			if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
				src.Close()
				c.JSON(http.StatusBadRequest, gin.H{"error": "foto tidak dapat dibaca"})
				return
			}
			if err := src.Close(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "foto tidak dapat dibaca"})
				return
			}
			typ := http.DetectContentType(head[:n])
			if typ != "image/jpeg" && typ != "image/png" {
				c.JSON(400, gin.H{"error": "isi file bukan gambar JPG/PNG"})
				return
			}
			contentTypes[i] = typ
		}
		if len(photos) > 0 {
			storage, err := NewR2StorageFromEnv(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "penyimpanan foto belum siap: " + err.Error()})
				return
			}
			for i, f := range photos {
				b := make([]byte, 16)
				if _, err := rand.Read(b); err != nil {
					cleanupUploadedReviewPhotos(c.Request.Context(), urls)
					c.JSON(500, gin.H{"error": "gagal menyimpan foto"})
					return
				}
				ext := strings.ToLower(filepath.Ext(f.Filename))
				name := hex.EncodeToString(b) + ext
				src, err := f.Open()
				if err != nil {
					cleanupUploadedReviewPhotos(c.Request.Context(), urls)
					c.JSON(http.StatusBadRequest, gin.H{"error": "foto tidak dapat dibaca"})
					return
				}
				photoURL, uploadErr := storage.Upload(c.Request.Context(), "ulasan/"+name, src, contentTypes[i])
				closeErr := src.Close()
				if uploadErr == nil {
					urls = append(urls, photoURL)
				}
				if uploadErr != nil || closeErr != nil {
					cleanupUploadedReviewPhotos(c.Request.Context(), urls)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "foto gagal disimpan ke penyimpanan"})
					return
				}
			}
		}
	}
	review, oldPhotoURLs, err := h.repo.Submit(digest(c.Param("token")), rating, text, urls)
	if err != nil {
		cleanupUploadedReviewPhotos(c.Request.Context(), urls)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := removeReviewPhotos(c.Request.Context(), oldPhotoURLs); err != nil {
		log.Printf("old review photo deletion failed after review submit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ulasan tersimpan, tetapi foto lama gagal dihapus"})
		return
	}
	c.JSON(201, gin.H{"message": "ulasan berhasil dikirim", "data": review})
}

func (h *Handler) list(c *gin.Context) {
	data, err := h.repo.ListForAdmin()
	if err != nil {
		c.JSON(500, gin.H{"error": "gagal mengambil ulasan"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *Handler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id ulasan tidak valid"})
		return
	}
	var input struct {
		Rating      int    `json:"rating"`
		KritikSaran string `json:"kritik_saran"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data ulasan tidak valid"})
		return
	}
	input.KritikSaran = strings.TrimSpace(input.KritikSaran)
	if input.Rating < 1 || input.Rating > 5 || input.KritikSaran == "" || len([]rune(input.KritikSaran)) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating harus 1–5 dan komentar wajib diisi (maksimal 5000 karakter)"})
		return
	}
	review, err := h.repo.UpdateReview(uint(id), input.Rating, input.KritikSaran)
	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ulasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memperbarui ulasan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": review})
}

func (h *Handler) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id ulasan tidak valid"})
		return
	}
	photoURLs, err := h.repo.DeleteReview(uint(id))
	if err != nil {
		if errors.Is(err, ErrReviewNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ulasan tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus ulasan"})
		return
	}
	if err := removeReviewPhotos(c.Request.Context(), photoURLs); err != nil {
		log.Printf("review photo deletion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ulasan dihapus, tetapi foto gagal dihapus"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ulasan berhasil dihapus"})
}

func removeReviewPhotos(ctx context.Context, urls []string) error {
	var storage *R2Storage
	for _, photoURL := range urls {
		const prefix = "/uploads/ulasan/"
		if strings.HasPrefix(photoURL, prefix) {
			name := strings.TrimPrefix(photoURL, prefix)
			if name == "" || filepath.Base(name) != name {
				return errors.New("nama file foto tidak valid")
			}
			if err := os.Remove(filepath.Join("uploads", "ulasan", name)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			continue
		}
		if storage == nil {
			var err error
			storage, err = NewR2StorageFromEnv(ctx)
			if err != nil {
				return err
			}
		}
		key, ok := storage.ObjectKey(photoURL)
		if !ok {
			return errors.New("url foto tidak valid")
		}
		if err := storage.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func cleanupUploadedReviewPhotos(ctx context.Context, urls []string) {
	if err := removeReviewPhotos(ctx, urls); err != nil {
		log.Printf("review photo cleanup failed: %v", err)
	}
}

func (h *Handler) publicList(c *gin.Context) {
	data, err := h.repo.List()
	if err != nil {
		c.JSON(500, gin.H{"error": "gagal mengambil ulasan"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}
