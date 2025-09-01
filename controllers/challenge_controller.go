package controllers

import (
	"net/http"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateChallenge handles the creation of a new challenge.
func CreateChallenge(c *gin.Context) {
	var input models.Challenge
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	// Optional: Validasi atau pengecekan tambahan jika dibutuhkan
	// Example: check if Vendor exists
	if err := config.DB.Where("id = ?", input.VendorID).First(&models.Vendor{}).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create challenge")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

// GetChallenges retrieves all the challenges.
func GetChallenges(c *gin.Context) {
	var challenges []models.Challenge
	if err := config.DB.Find(&challenges).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch challenges")
		return
	}
	response.JSONSuccess(c.Writer, true, http.StatusOK, challenges)
}

func CreateChallengeLog(c *gin.Context) {
	var input models.ChallengeLog
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	// Cek user
	var user models.User
	if err := config.DB.First(&user, input.UserID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "User not found")
		return
	}

	// Cek challenge
	var challenge models.Challenge
	if err := config.DB.First(&challenge, input.ChallengeID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Challenge not found")
		return
	}

	// Optional: validasi vendor cocok
	if user.VendorID != challenge.VendorID {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "User and challenge are not in the same vendor")
		return
	}

	// Set vendor ID ke log (jika model ChallengeLog punya field VendorID)
	input.VendorID = challenge.VendorID

	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create challenge log")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

// GetChallengeLogs retrieves all the challenge logs.
// func GetChallengeLogs(c *gin.Context) {
// 	var challengeLogs []models.ChallengeLog
// 	if err := config.DB.Find(&challengeLogs).Error; err != nil {
// 		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch challenge logs")
// 		return
// 	}
// 	response.JSONSuccess(c.Writer, true, http.StatusOK, challengeLogs)
// }

func GetChallengesByVendor(c *gin.Context) {
	vendorIDStr := c.Query("vendor_id")
	if vendorIDStr == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Vendor ID is required")
		return
	}

	vendorID, err := strconv.ParseUint(vendorIDStr, 10, 64)
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid vendor ID")
		return
	}

	var challenges []models.Challenge
	if err := config.DB.Where("vendor_id = ?", vendorID).Find(&challenges).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch challenges")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, challenges)
}

func GetChallengeLogs(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := uint(userIDRaw.(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User not found")
		return
	}

	var challengeLogs []models.ChallengeLog
	query := config.DB.Model(&models.ChallengeLog{})

	switch user.Role {
	case "member":
		query = query.Where("user_id = ?", user.ID)
	case "pelatih":
		// ambil semua user di vendor yang sama
		var userIDs []uint
		config.DB.Model(&models.User{}).
			Where("vendor_id = ?", user.VendorID).
			Pluck("id", &userIDs)
		query = query.Where("user_id IN ?", userIDs)
	}

	if err := query.Find(&challengeLogs).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch challenge logs")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, challengeLogs)
}
