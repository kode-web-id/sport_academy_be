package controllers

import (
	"net/http"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateEvent handles the creation of a new event.
func CreateEvent(c *gin.Context) {
	var input models.Event
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := config.DB.Where("id = ?", input.VendorID).First(&models.Vendor{}).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}
	input.IsFinish = false

	// Menyimpan Event baru
	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create event")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}

func GetEvents(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	eventType := c.DefaultQuery("event_type", "")
	isPaid := c.DefaultQuery("is_paid", "")
	id := c.DefaultQuery("id", "")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	userIDRaw, _ := c.Get("user_id")
	userID := uint(userIDRaw.(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User not found")
		return
	}

	var events []models.Event
	var totalEvents int64

	query := config.DB.Model(&models.Event{}).Preload("Users")

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if id != "" {
		query = query.Where("id = ?", id)
	}
	if isPaid != "" {
		isPaidBool, _ := strconv.ParseBool(isPaid)
		query = query.Where("is_paid = ?", isPaidBool)
	}

	query = query.Where("vendor_id = ?", user.VendorID)

	query.Count(&totalEvents)
	query.Offset((pageInt - 1) * limitInt).Limit(limitInt).Find(&events)

	pagination := gin.H{
		"page":  pageInt,
		"limit": limitInt,
		"total": totalEvents,
	}
	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"pagination": pagination,
		"events":     events,
	})
}

func UpdateEventFinishStatus(c *gin.Context) {
	type FinishUpdateInput struct {
		ID       uint `json:"id"`
		IsFinish bool `json:"is_finish"`
	}

	var input FinishUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	var event models.Event
	if err := config.DB.First(&event, input.ID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event not found")
		return
	}

	event.IsFinish = input.IsFinish

	if err := config.DB.Save(&event).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update event status")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, event)
}

func CreateEventLog(c *gin.Context) {
	var input models.EventLog
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	// Pastikan Vendor dan Event ada
	if err := config.DB.Where("id = ?", input.VendorID).First(&models.Vendor{}).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}
	var event models.Event
	if err := config.DB.First(&event, input.EventID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event not found")
		return
	}

	// Cek apakah EventLog dengan kombinasi UserID dan EventID sudah ada
	var existing models.EventLog
	if err := config.DB.Where("user_id = ? AND event_id = ?", input.UserID, input.EventID).First(&existing).Error; err == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Event log already exists for this user and event")
		return
	}

	// Menyimpan EventLog baru
	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create event log")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, input)
}
func UpdateEventLogStatus(c *gin.Context) {
	// Define struct for request body
	var input struct {
		ID     uint   `json:"id"`     // ID dari event log yang akan diupdate
		Status bool   `json:"status"` // Status baru untuk event log
		Note   string `json:"note"`
	}

	// Bind the request body to input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	// Fetch event log based on the ID from input
	var eventLog models.EventLog
	if err := config.DB.First(&eventLog, input.ID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event log not found")
		return
	}

	// Update the status of the event log
	eventLog.Status = input.Status
	eventLog.Note = input.Note

	// Save the updated event log to the database
	if err := config.DB.Save(&eventLog).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update status")
		return
	}

	// Return the updated event log
	response.JSONSuccess(c.Writer, true, http.StatusOK, eventLog)
}

func GetEventLogs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	eventType := c.DefaultQuery("event_type", "")
	status := c.DefaultQuery("status", "")
	eventID := c.DefaultQuery("event_id", "")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	userIDRaw, _ := c.Get("user_id")
	userID := uint(userIDRaw.(float64))

	// Ambil data user
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User not found")
		return
	}

	var eventLogs []models.EventLog
	var totalLogs int64

	query := config.DB.Model(&models.EventLog{}).Where("vendor_id = ?", user.VendorID)

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	query.Count(&totalLogs)
	query.Offset((pageInt - 1) * limitInt).Limit(limitInt).Find(&eventLogs)

	pagination := gin.H{
		"page":  pageInt,
		"limit": limitInt,
		"total": totalLogs,
	}
	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"pagination": pagination,
		"event_logs": eventLogs,
	})
}

func GetEventLogsByUser(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	eventType := c.DefaultQuery("event_type", "")
	status := c.DefaultQuery("status", "")
	eventID := c.DefaultQuery("event_id", "")

	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)

	userIDRaw, _ := c.Get("user_id")
	userID := uint(userIDRaw.(float64))

	// Cek apakah user ada
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User not found")
		return
	}

	var eventLogs []models.EventLog
	var totalLogs int64

	query := config.DB.Model(&models.EventLog{}).Where("user_id = ?", userID)

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	query.Count(&totalLogs)
	query.Offset((pageInt - 1) * limitInt).Limit(limitInt).Find(&eventLogs)

	pagination := gin.H{
		"page":  pageInt,
		"limit": limitInt,
		"total": totalLogs,
	}
	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"pagination": pagination,
		"event_logs": eventLogs,
	})
}
