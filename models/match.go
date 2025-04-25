package models

import "gorm.io/gorm"

type Match struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	VendorID    *uint  `json:"vendor_id"`
	// Vendor      *Vendor `gorm:"foreignKey:VendorID"`
	EventID *uint `json:"event_id"`
	// Event       *Event  `gorm:"foreignKey:EventID"`
}
