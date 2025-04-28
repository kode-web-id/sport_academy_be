package controllers

import (
	"net/http"
	"regexp"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"ssb_api/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Cek format email
	emailRegex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, input.Email)
	if err != nil || !matched {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Cek apakah email sudah terdaftar
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Email sudah terdaftar")
		return
	}

	// Cek apakah telephone sudah terdaftar
	if input.Phone != "" {
		var existingPhone models.User
		if err := config.DB.Where("phone = ?", input.Phone).First(&existingPhone).Error; err == nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Nomor telepon sudah terdaftar")
			return
		}
	}

	// Hash password
	hashed, err := utils.HasingPassword(input.Password)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	input.Password = hashed

	// Fetch vendor if ada
	if input.VendorID != 0 {
		var vendor models.Vendor
		if err := config.DB.First(&vendor, input.VendorID).Error; err != nil {
			response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
			return
		}
		input.Vendor = vendor
	}

	// Create user
	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create user")
		return
	}

	input.Password = ""

	// Success response
	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Invalid email")
		return
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Menghilangkan password dari response
	user.Password = "" // Atau bisa menggunakan `user.Password = nil` jika tipe pointer
	baseURL := utils.DotEnv("BASE_URL_F")

	if user.Photo != "" {
		// Bangun URL foto menggunakan baseURL yang didapat dari .env
		user.Photo = baseURL + "/" + strings.TrimPrefix(user.Photo, "./")
	}
	result := gin.H{
		"user": gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"role":      user.Role,
			"phone":     user.Phone,
			"address":   user.Address,
			"gender":    user.Gender,
			"birthDate": user.BirthDate,
			"photo":     user.Photo,
			"vendor_id": user.VendorID,
		},
		"token": token,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}
