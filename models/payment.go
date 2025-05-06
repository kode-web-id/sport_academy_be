package models

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	UserID   uint    `json:"user_id"`
	VendorID uint    `json:"vendor_id"`
	EventID  *uint   `json:"event_id"` // nullable
	Amount   float64 `json:"amount"`
	Method   string  `json:"method"` // cash, transfer, e-wallet
	Status   string  `json:"status"` // pending, success, failed
	Type     string  `json:"type"`   // general, event
	Date     string  `json:"date"`
	Note     string  `json:"note"`
	Photo    string  `json:"photo"`
	Invoice  string  `json:"invoice"` // ← new column
	UserName string  `json:"user_name"`
}

type PaymentRequest struct {
	UserID   uint    `form:"user_id"`
	VendorID uint    `form:"vendor_id"`
	EventID  *uint   `form:"event_id"`
	Amount   float64 `form:"amount"`
	Method   string  `form:"method"`
	Status   string  `form:"status"`
	Type     string  `form:"type"`
	Date     string  `form:"date"`
	Note     string  `form:"note"`
	UserName string  `json:"user_name"`
	Invoice  string  `form:"invoice"` // ← new field
}
