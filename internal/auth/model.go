package auth

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nama         string `json:"nama"`
	Email        string `json:"email" gorm:"unique"`
	PasswordHash string `json:"-"`
	Role         string `json:"role" gorm:"default:owner"`
}
