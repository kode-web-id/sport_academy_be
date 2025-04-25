package models

import "gorm.io/gorm"

type EventLog struct {
	gorm.Model
	UserID   uint   `json:"user_id"`
	EventID  uint   `json:"event_id"`
	VendorID uint   `json:"vendor_id"`
	UserName string `json:"user_name"`

	// ID event-nya
	EventType string `json:"event_type"` // contoh: "training", "match", "tournament", "challenge"
	Note      string `json:"note"`
	Status    bool   `json:"status"`
	// User      User   `gorm:"foreignKey:UserID"`
	// Vendor    Vendor `gorm:"foreignKey:VendorID"`
}
