package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string `json:"name"`
	Email     string `json:"email" gorm:"unique"`
	Phone     string `json:"phone" gorm:"unique"`
	Password  string `json:"password"`
	Role      string `json:"role"` // "pemain", "pelatih", "admin"
	Address   string `json:"address"`
	Photo     string `json:"photo"`
	Gender    string `json:"gender"`     // "Laki-laki", "Perempuan"
	BirthDate string `json:"birth_date"` // format: YYYY-MM-DD
	Status    string `json:"status"`     // "free" atau "pro"
	VendorID  uint   `json:"vendor_id"`
	Vendor    Vendor `gorm:"foreignKey:VendorID"`
	// Payments  []Payment `gorm:"foreignKey:UserID"`
}
