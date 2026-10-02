package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sewa-jas-api/config"
	"sewa-jas-api/internal/auth"
	"sewa-jas-api/internal/barang"
	"sewa-jas-api/internal/denda"
	"sewa-jas-api/internal/notifikasi"
	"sewa-jas-api/internal/pengeluaran"
	"sewa-jas-api/internal/penyewa"
	"sewa-jas-api/internal/transaksi"
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
	)

	router := gin.Default()
	router.Use(config.CORS())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth.RegisterRoutes(router, auth.NewRepository(db))

	protected := router.Group("/api")
	protected.Use(auth.Middleware())

	barangRepo := barang.NewRepository(db)
	transaksiRepo := transaksi.NewRepository(db)

	barang.RegisterRoutes(protected, barangRepo)
	penyewa.RegisterRoutes(protected, penyewa.NewRepository(db))
	transaksi.RegisterRoutes(protected, transaksiRepo)
	denda.RegisterRoutes(protected, denda.NewRepository(db))
	pengeluaran.RegisterRoutes(protected, pengeluaran.NewRepository(db))
	notifikasi.RegisterRoutes(protected, transaksiRepo)

	router.Run(":8080")
}
