package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	Title    string `json:"title"`
	Category string `json:"category"` // contoh: stamina, passing, shooting
	MaxPoint int    `json:"max_point"`
	VendorID *uint  `json:"vendor_id"`
	Vendor   Vendor `gorm:"foreignKey:VendorID"`
	EventID  *uint  `json:"event_id"`
	// Event    *Event `gorm:"foreignKey:EventID"`
}
