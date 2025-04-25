package models

import "gorm.io/gorm"

type ChallengeLog struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	ChallengeID uint    `json:"challenge_id"`
	Point       float64 `json:"point"` // Poin yang didapatkan oleh user
	Note        string  `json:"note"`  // Catatan tambahan jika ada
	VendorID    uint    `json:"vendor_id"`
	Vendor      Vendor  `gorm:"foreignKey:VendorID"`
	// User        User      `gorm:"foreignKey:UserID"`
	// Challenge   Challenge `gorm:"foreignKey:ChallengeID"`
}
