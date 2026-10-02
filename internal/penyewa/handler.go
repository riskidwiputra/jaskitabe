package penyewa

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo *Repository }

func RegisterRoutes(router gin.IRouter, repo *Repository) {
	h := &Handler{repo: repo}
	group := router.Group("/penyewa")
	group.GET("", h.list)
	group.POST("", h.create)
}

func (h *Handler) list(c *gin.Context) {
	data, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *Handler) create(c *gin.Context) {
	var input Penyewa
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Create(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": input})
}
