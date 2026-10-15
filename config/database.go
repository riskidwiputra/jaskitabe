package config

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak menemukan file .env, memakai environment variable yang sudah ada")
	}
	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	if host == "" || port == "" {
		log.Fatal("DB_HOST dan DB_PORT wajib dikonfigurasi")
	}

	driverConfig := mysqldriver.Config{
		User:                 os.Getenv("DB_USER"),
		Passwd:               os.Getenv("DB_PASSWORD"),
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(host, port),
		DBName:               os.Getenv("DB_NAME"),
		ParseTime:            true,
		Loc:                  time.Local,
		Params:               map[string]string{"charset": "utf8mb4"},
		AllowNativePasswords: true,
	}
	if strings.EqualFold(os.Getenv("DB_TLS"), "true") {
		if err := mysqldriver.RegisterTLSConfig("tidb", &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		}); err != nil {
			log.Fatalf("Gagal menyiapkan TLS database: %v", err)
		}
		driverConfig.TLSConfig = "tidb"
	}

	db, err := gorm.Open(gormmysql.Open(driverConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}
	log.Println("Berhasil konek ke database")
	return db
}
