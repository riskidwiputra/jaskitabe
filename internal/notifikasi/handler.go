package notifikasi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sewa-jas-api/internal/transaksi"
)

type Handler struct{ transaksiRepo *transaksi.Repository }

func RegisterRoutes(router gin.IRouter, transaksiRepo *transaksi.Repository) {
	h := &Handler{transaksiRepo: transaksiRepo}
	router.GET("/notifikasi/jatuh-tempo", h.jatuhTempo)
}

func (h *Handler) jatuhTempo(c *gin.Context) {
	data, err := h.transaksiRepo.FindJatuhTempo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "jumlah": len(data)})
}
