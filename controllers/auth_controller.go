package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"ssb_api/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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
	if input.VendorID != nil {
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
		FCMToken string `json:"fcm_token"`
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

	if input.FCMToken != "" {
		config.DB.Model(&user).Update("fcm_token", input.FCMToken)
	}

	// ✅ Generate access + refresh token
	accessToken, refreshToken, err := utils.GenerateTokens(user.ID, user.Email)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	user.Password = ""
	baseURL := utils.DotEnv("BASE_URL_F")
	if user.Photo != "" {
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
		"token":         accessToken,
		"refresh_token": refreshToken,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}

func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Missing refresh token")
		return
	}

	// ✅ Parse refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil || claims["email"] == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Invalid claims")
		return
	}

	// ✅ Generate access token baru
	accessToken, _, err := utils.GenerateTokens(uint(claims["user_id"].(float64)), claims["email"].(string))
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func FirebaseLogin(c *gin.Context) {
	var input struct {
		IDToken  string `json:"id_token" binding:"required"`
		FCMToken string `json:"fcm_token"`
	}

	// 🔹 Validasi input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Missing token")
		return
	}

	ctx := context.Background()

	// 🔹 Inisialisasi Firebase Auth
	client, err := config.FirebaseApp.Auth(ctx)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Firebase not initialized")
		return
	}

	// 🔹 Verifikasi ID token Firebase
	token, err := client.VerifyIDToken(ctx, input.IDToken)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Invalid ID token")
		return
	}

	// 🔹 Ambil email
	emailClaim, ok := token.Claims["email"]
	if !ok {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Email not found in token")
		return
	}
	email := emailClaim.(string)

	// 🔹 Ambil nama & foto jika ada
	name := ""
	if val, ok := token.Claims["name"]; ok {
		name = val.(string)
	}

	photo := ""
	if val, ok := token.Claims["picture"]; ok {
		photo = val.(string)
	}

	// 🔹 Cek apakah user sudah ada
	var user models.User
	err = config.DB.First(&user, "email = ?", email).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// User belum ada → FE harus register
		response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
			"need_register": true,
			"email":         email,
			"name":          name,
			"photo":         photo,
		})
		return
	}

	// 🔹 Update FCM token jika ada
	if input.FCMToken != "" {
		config.DB.Model(&user).Update("fcm_token", input.FCMToken)
	}

	// 🔹 Generate JWT
	accessToken, refreshToken, err := utils.GenerateTokens(user.ID, user.Email)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create token")
		return
	}

	// 🔹 Bersihkan password dari response
	user.Password = ""
	baseURL := utils.DotEnv("BASE_URL_F")
	if user.Photo != "" {
		user.Photo = baseURL + "/" + strings.TrimPrefix(user.Photo, "./")
	}

	// 🔹 Cek apakah user harus lengkapi profil
	mustCompleteProfile := user.VendorID == nil || *user.VendorID == 0

	// 🔹 Response sama dengan login biasa
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
		"token":                 accessToken,
		"refresh_token":         refreshToken,
		"must_complete_profile": mustCompleteProfile,
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, result)
}

func RedirectToApp(c *gin.Context) {
	id := c.Param("id")
	userAgent := c.GetHeader("User-Agent")

	// ✅ Deteksi platform user
	isAndroid := strings.Contains(strings.ToLower(userAgent), "android")
	isIOS := strings.Contains(strings.ToLower(userAgent), "iphone") || strings.Contains(strings.ToLower(userAgent), "ipad")

	// URL Play Store / App Store fallback
	playStoreURL := "https://play.google.com/store/apps/details?id=one.team"
	appStoreURL := "https://apps.apple.com/app/idYOUR_APP_ID" // opsional

	// Intent scheme (khusus Android)
	intentURL := fmt.Sprintf(
		"intent://event/%s#Intent;scheme=oneteam;package=one.team;S.browser_fallback_url=%s;end",
		id, playStoreURL,
	)

	if isAndroid {
		// 🔁 Redirect ke intent:// biar Android langsung buka app
		c.Redirect(http.StatusFound, intentURL)
		return
	}

	if isIOS {
		// 🍎 iOS redirect ke Universal Link (kalau sudah setup)
		// kalau belum, fallback ke App Store
		c.Redirect(http.StatusFound, appStoreURL)
		return
	}

	// 🌐 fallback umum (PC, browser, dsb)
	c.String(http.StatusOK, fmt.Sprintf(`
		<html>
		<head>
			<meta http-equiv="refresh" content="0; url=%s" />
		</head>
		<body>
			<p>Jika tidak terbuka otomatis, klik di sini: <a href="%s">Buka di aplikasi OneTeam</a></p>
		</body>
		</html>
	`, intentURL, intentURL))
}
