package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"ssb_api/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateVendor(c *gin.Context) {
	var input models.Vendor
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	emailRegex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, input.Email)
	if err != nil || !matched {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid email format")
		return
	}

	var existingEmail models.Vendor
	if err := config.DB.Where("email = ?", input.Email).First(&existingEmail).Error; err == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusConflict, "Email already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Error checking email")
		return
	}

	var existingPhone models.Vendor
	if err := config.DB.Where("phone = ?", input.Phone).First(&existingPhone).Error; err == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusConflict, "Phone number already exists")
		return
	} else if err != gorm.ErrRecordNotFound {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Error checking phone number")
		return
	}
	if input.Name == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Nama wajib diisi")
		return
	}

	if input.Email == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Email wajib diisi")
		return
	}

	if input.Phone == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Nomor telepon wajib diisi")
		return
	}

	if input.Address == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Alamat wajib diisi")
		return
	}

	if input.Category == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Kategori wajib diisi")
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create academy")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

func GetVendors(c *gin.Context) {
	var vendors []models.Vendor
	name := c.Query("name") // menangkap parameter ?name=...

	query := config.DB.Model(&models.Vendor{})
	if name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	if err := query.Find(&vendors).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Gagal mengambil data vendor")
		return
	}

	// Base URL untuk gambar vendor
	baseURL := utils.DotEnv("BASE_URL_F")
	baseURL = strings.TrimRight(baseURL, "/") + "/"

	// Menambahkan base URL ke foto vendor
	for i := range vendors {
		// Jika foto vendor tidak kosong, hilangkan prefix './' dan gabungkan dengan base URL
		if vendors[i].Photo != "" {
			vendors[i].Photo = baseURL + strings.TrimPrefix(vendors[i].Photo, "./")
		}
	}

	// Mengembalikan data vendor dengan foto yang sudah lengkap dengan base URL
	response.JSONSuccess(c.Writer, true, http.StatusOK, vendors)
}

func UpdateVendorPhoto(c *gin.Context) {
	// Ambil ID Vendor dari parameter URL
	vendorID := c.Param("id")

	// Ambil file foto yang diupload
	file, err := c.FormFile("photo")
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "No file is attached")
		return
	}

	// Ambil vendor dari database
	var vendor models.Vendor
	if err := config.DB.First(&vendor, vendorID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(vendor.Name, " ", "_"))
	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().Unix()

	// Tentukan nama file berdasarkan vendor ID, slug, dan timestamp
	dst := fmt.Sprintf("./uploads/vendors/%d_%s_%d%s", vendor.ID, slug, timestamp, ext)

	// Simpan file ke server
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Update foto vendor
	vendor.Photo = dst
	if err := config.DB.Save(&vendor).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update vendor photo")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, "Update foto succesfully")
}

func UpdateVendorBank(c *gin.Context) {
	var input struct {
		BankName      string `json:"bank_name"`
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
	}

	// Binding request body
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get UserID from JWT token
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	userID := uint(userIDRaw.(float64)) // default JWT parsing returns float64

	// Retrieve user details from database
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "User not found")
		return
	}

	// Only allow trainers (pelatih) to update bank information
	if user.Role != "pelatih" {
		response.JSONErrorResponse(c.Writer, false, http.StatusForbidden, "Only trainers can update vendor bank information")
		return
	}

	// Retrieve vendor associated with the logged-in trainer
	var vendor models.Vendor
	if err := config.DB.First(&vendor, user.VendorID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}

	// Update vendor bank details
	vendor.BankName = input.BankName
	vendor.BankAccount = input.AccountName
	vendor.BankHolder = input.AccountNumber

	if err := config.DB.Save(&vendor).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update vendor bank information")
		return
	}

	// Return success response
	response.JSONSuccess(c.Writer, true, http.StatusOK, "Update bank succesfully")
}
