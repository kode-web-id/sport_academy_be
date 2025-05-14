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
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUserFromToken(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(float64) // JWT MapClaims returns float64 for numbers
	if !ok {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Invalid user ID format")
		return
	}

	var user models.User
	if err := config.DB.First(&user, uint(userID)).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "User not found")
		return
	}
	if user.VendorID != 0 {
		var vendor models.Vendor
		if err := config.DB.First(&vendor, user.VendorID).Error; err != nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
			return
		}
		// Set the Vendor in the User struct (optional)
		user.Vendor = vendor
	}

	baseURL := utils.DotEnv("BASE_URL_F")

	if user.Photo != "" {
		// Bangun URL foto menggunakan baseURL yang didapat dari .env
		user.Photo = baseURL + "/" + strings.TrimPrefix(user.Photo, "./")
	}
	if user.Vendor.Photo != "" {
		// Bangun URL foto menggunakan baseURL yang didapat dari .env
		user.Vendor.Photo = baseURL + "/" + strings.TrimPrefix(user.Vendor.Photo, "./")
	}

	// Hilangkan password dari response
	user.Password = ""

	response.JSONSuccess(c.Writer, true, http.StatusOK, user)
}

func SearchUsers(c *gin.Context) {
	var users []models.User
	var total int64

	// Ambil page & limit
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, err1 := strconv.Atoi(pageStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil || page < 1 || limit < 1 {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid page or limit")
		return
	}

	offset := (page - 1) * limit

	query := config.DB.Model(&models.User{})

	// Filter berdasarkan parameter query
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if email := c.Query("email"); email != "" {
		query = query.Where("email ILIKE ?", "%"+email+"%")
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// Hitung total yang sesuai filter
	if err := query.Count(&total).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to count filtered users")
		return
	}

	// Ambil data
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to get filtered users")
		return
	}
	userResponses := make([]gin.H, len(users))
	baseURL := utils.DotEnv("BASE_URL_F")
	baseURL = strings.TrimRight(baseURL, "/") + "/" // Pastikan trailing slash ada

	for i, user := range users {
		// Jika foto tidak kosong, tambahkan base URL
		if user.Photo != "" {
			user.Photo = baseURL + strings.TrimPrefix(user.Photo, "./")
		}

		userResponses[i] = gin.H{
			"id":           user.ID,
			"name":         user.Name,
			"email":        user.Email,
			"role":         user.Role,
			"phone":        user.Phone,
			"address":      user.Address,
			"gender":       user.Gender,
			"birthDate":    user.BirthDate,
			"status":       user.Status,
			"vendor_id":    user.VendorID,
			"vendor":       user.Vendor,
			"photo":        user.Photo,
			"position":     user.Position,
			"foot":         user.Foot,
			"number":       user.Number,
			"age_category": user.AgeCategory,
			"star":         user.Star,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		}

	}

	result := gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"users": userResponses,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}

func GetAllUsers(c *gin.Context) {
	var users []models.User
	var total int64

	// Ambil query param: page & limit
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	// Konversi ke int
	page, err1 := strconv.Atoi(pageStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil || page < 1 || limit < 1 {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid page or limit")
		return
	}

	offset := (page - 1) * limit

	db := config.DB.Model(&models.User{})

	// Hitung total
	if err := db.Count(&total).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to count users")
		return
	}

	// Ambil data
	if err := db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to get users")
		return
	}

	userResponses := make([]gin.H, len(users))
	baseURL := utils.DotEnv("BASE_URL_F")
	baseURL = strings.TrimRight(baseURL, "/") + "/" // Pastikan trailing slash ada

	for i, user := range users {
		// Jika foto tidak kosong, tambahkan base URL
		if user.Photo != "" {
			user.Photo = baseURL + strings.TrimPrefix(user.Photo, "./")
		}

		userResponses[i] = gin.H{
			"id":           user.ID,
			"name":         user.Name,
			"email":        user.Email,
			"role":         user.Role,
			"phone":        user.Phone,
			"address":      user.Address,
			"gender":       user.Gender,
			"birthDate":    user.BirthDate,
			"status":       user.Status,
			"vendor_id":    user.VendorID,
			"vendor":       user.Vendor,
			"photo":        user.Photo,
			"position":     user.Position,
			"foot":         user.Foot,
			"number":       user.Number,
			"age_category": user.AgeCategory,
			"star":         user.Star,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		}

	}

	result := gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"users": userResponses,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}

func UpdateUserPhoto(c *gin.Context) {
	// Ambil ID User dari token atau parameter
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	userID := uint(userIDRaw.(float64))

	// Ambil file foto yang diupload
	file, err := c.FormFile("photo")
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "No file is attached")
		return
	}

	// Ambil user dari database
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "User not found")
		return
	}

	// Tentukan nama file berdasarkan user ID atau user name
	slug := strings.ToLower(strings.ReplaceAll(user.Name, " ", "_"))
	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().Unix()

	// Tentukan nama file berdasarkan user ID, slug, dan timestamp
	dst := fmt.Sprintf("./uploads/users/%d_%s_%d%s", user.ID, slug, timestamp, ext)

	// Simpan file ke server
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Update foto user
	user.Photo = dst
	if err := config.DB.Save(&user).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update user photo")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, "Update foto succesfully")
}

func GetUsersByVendor(c *gin.Context) {
	var users []models.User
	var total int64

	// Ambil query param: page & limit
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	vendorIDStr := c.DefaultQuery("vendor_id", "")

	// Konversi page dan limit ke int
	page, err1 := strconv.Atoi(pageStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil || page < 1 || limit < 1 {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid page or limit")
		return
	}

	// Validasi vendor_id
	if vendorIDStr == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Vendor ID is required")
		return
	}

	vendorID, err := strconv.Atoi(vendorIDStr)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid vendor ID")
		return
	}

	offset := (page - 1) * limit

	// Query untuk mengambil data users berdasarkan VendorID
	db := config.DB.Model(&models.User{}).Where("vendor_id = ?", vendorID)

	// Hitung total users
	if err := db.Count(&total).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to count users")
		return
	}

	// Ambil data users
	if err := db.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to get users")
		return
	}

	userResponses := make([]gin.H, len(users))
	baseURL := utils.DotEnv("BASE_URL_F")
	baseURL = strings.TrimRight(baseURL, "/") + "/" // Pastikan trailing slash ada

	for i, user := range users {
		// Jika foto tidak kosong, tambahkan base URL
		if user.Photo != "" {
			user.Photo = baseURL + strings.TrimPrefix(user.Photo, "./")
		}

		userResponses[i] = gin.H{
			"id":           user.ID,
			"name":         user.Name,
			"email":        user.Email,
			"role":         user.Role,
			"phone":        user.Phone,
			"address":      user.Address,
			"gender":       user.Gender,
			"birthDate":    user.BirthDate,
			"status":       user.Status,
			"vendor_id":    user.VendorID,
			"vendor":       user.Vendor,
			"photo":        user.Photo,
			"position":     user.Position,
			"foot":         user.Foot,
			"number":       user.Number,
			"age_category": user.AgeCategory,
			"star":         user.Star,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		}

	}

	// Buat response
	result := gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"users": userResponses,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}

func UpdateUser(c *gin.Context) {
	var input models.User

	// Bind body JSON ke struct input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	if input.ID == 0 {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "User ID wajib diisi")
		return
	}

	// Ambil user berdasarkan ID
	var user models.User
	if err := config.DB.First(&user, input.ID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "User not found")
		return
	}

	// Cek format email jika diubah
	if input.Email != "" && input.Email != user.Email {
		emailRegex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
		matched, err := regexp.MatchString(emailRegex, input.Email)
		if err != nil || !matched {
			response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid email format")
			return
		}
		var existingUser models.User
		if err := config.DB.Where("email = ?", input.Email).Not("id = ?", input.ID).First(&existingUser).Error; err == nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Email sudah terdaftar")
			return
		}
		user.Email = input.Email
	}

	// Cek nomor telepon jika diubah
	if input.Phone != "" && input.Phone != user.Phone {
		var existingPhone models.User
		if err := config.DB.Where("phone = ?", input.Phone).Not("id = ?", input.ID).First(&existingPhone).Error; err == nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Nomor telepon sudah terdaftar")
			return
		}
		user.Phone = input.Phone
	}

	// Jika password diubah
	if input.Password != "" {
		hashed, err := utils.HasingPassword(input.Password)
		if err != nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		user.Password = hashed
	}

	// Update nama, role
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Role != "" {
		user.Role = input.Role
	}

	// Update vendor jika berbeda
	if input.VendorID != 0 && input.VendorID != user.VendorID {
		var vendor models.Vendor
		if err := config.DB.First(&vendor, input.VendorID).Error; err != nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
			return
		}
		user.VendorID = vendor.ID
	}

	// Simpan perubahan
	if err := config.DB.Save(&user).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update user")
		return
	}

	user.Password = ""
	response.JSONSuccess(c.Writer, true, http.StatusOK, user)
}
