package models

import "gorm.io/gorm"

type Vendor struct {
	gorm.Model
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Photo       string `json:"photo"`
	BankName    string `json:"bank_name"`    // Nama bank, contoh: BCA, Mandiri
	BankAccount string `json:"bank_account"` // Nomor rekening
	BankHolder  string `json:"bank_holder"`
	Category    string `json:"category"`
	// Payments    []Payment `gorm:"foreignKey:VendorID"`
}
