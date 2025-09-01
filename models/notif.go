package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"` // untuk siapa notifikasinya
	Title     string
	Body      string
	Type      string     // contoh: "tagihan", "status_pembayaran"
	Token     string     // device token user
	IsSent    bool       `gorm:"default:false"` // apakah sudah dikirim
	IsRead    bool       `gorm:"default:false"`
	SentAt    *time.Time // waktu pengiriman jika berhasil
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type FcmMessage struct {
	To           string            `json:"to"`
	Data         map[string]string `json:"data,omitempty"`
	Notification map[string]string `json:"notification,omitempty"`
}

type CreateNotificationInput struct {
	UserID uint   `json:"user_id" binding:"required"`
	Title  string `json:"title" binding:"required"`
	Body   string `json:"body" binding:"required"`
	Type   string `json:"type" binding:"required"`  // e.g., tagihan / status_pembayaran
	Token  string `json:"token" binding:"required"` // Device token
}
