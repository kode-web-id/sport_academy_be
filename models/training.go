package models

import "gorm.io/gorm"

type Training struct {
	gorm.Model
	UserID   uint
	Type     string `json:"type"` // fisik, teknik, taktik
	Notes    string `json:"notes"`
	VendorID *uint  // tambahkan ini untuk relasi opsional ke vendor
	// Vendor   *Vendor `gorm:"foreignKey:VendorID"`
	EventID *uint `json:"event_id"`
	// Event    *Event  `gorm:"foreignKey:EventID"`
}
