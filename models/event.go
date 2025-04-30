package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	EventType     string  `json:"event_type"` // match, tournament, challenge, etc.
	Date          string  `json:"date"`
	Time          string  `json:"time"`
	Location      string  `json:"location"`
	LocationPoint string  `json:"location_point"`
	IsPaid        bool    `json:"is_paid"`
	PaymentType   string  `json:"payment_type"` // daily, monthly
	Fee           float64 `json:"fee"`
	VendorID      uint    `json:"vendor_id"`
	IsFinish      bool    `json:"is_finish"`

	// Vendor        Vendor  `gorm:"foreignKey:VendorID"`
	// Users []User `gorm:"many2many:event_participants"`
}
