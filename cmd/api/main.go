package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"sewa-jas-api/config"
	"sewa-jas-api/internal/auth"
	"sewa-jas-api/internal/barang"
	"sewa-jas-api/internal/denda"
	"sewa-jas-api/internal/notifikasi"
	"sewa-jas-api/internal/pengeluaran"
	"sewa-jas-api/internal/penyewa"
	"sewa-jas-api/internal/transaksi"
	"sewa-jas-api/internal/ulasan"
)

func main() {
	db := config.Connect()

	db.AutoMigrate(
		&auth.User{},
		&barang.Barang{},
		&penyewa.Penyewa{},
		&transaksi.TransaksiSewa{},
		&transaksi.TransaksiSewaBarang{},
		&denda.Denda{},
		&pengeluaran.Pengeluaran{},
		&ulasan.Token{},
		&ulasan.Review{},
		&ulasan.ReviewPhoto{},
	)

	router := gin.Default()
	router.Use(config.CORS())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth.RegisterRoutes(router, auth.NewRepository(db))

	protected := router.Group("/api")
	protected.Use(auth.Middleware())
	public := router.Group("/api/public")

	barangRepo := barang.NewRepository(db)
	transaksiRepo := transaksi.NewRepository(db)

	ulasan.RegisterRoutes(public, protected, ulasan.NewRepository(db))
	barang.RegisterRoutes(protected, barangRepo)
	penyewa.RegisterRoutes(protected, penyewa.NewRepository(db))
	transaksi.RegisterRoutes(protected, transaksiRepo)
	denda.RegisterRoutes(protected, denda.NewRepository(db))
	pengeluaran.RegisterRoutes(protected, pengeluaran.NewRepository(db))
	notifikasi.RegisterRoutes(protected, transaksiRepo)

	router.Static("/uploads/ulasan", "./uploads/ulasan")
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan API: %v", err)
	}
}
