package controllers

import (
	"net/http"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateMatch(c *gin.Context) {
	var input models.Match
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}
	if err := config.DB.Where("id = ?", input.VendorID).First(&models.Vendor{}).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create training")
		return
	}
	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

func GetMatchs(c *gin.Context) {
	var trainings []models.Match
	if err := config.DB.Find(&trainings).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch trainings")
		return
	}
	response.JSONSuccess(c.Writer, true, http.StatusOK, trainings)
}

func GetMatchsByVendor(c *gin.Context) {
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

	var trainings []models.Match
	if err := config.DB.Where("vendor_id = ?", vendorID).Find(&trainings).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch matchs by vendor")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, trainings)
}
