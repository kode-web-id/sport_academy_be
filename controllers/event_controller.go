package controllers

import (
	"net/http"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"strconv"
	"strings"

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

func UpdateEvent(c *gin.Context) {
	// Ambil ID event dari URL param
	eventID := c.Param("id")
	var input models.Event

	// Bind JSON body
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request")
		return
	}

	// Cari event yang akan diupdate
	var event models.Event
	if err := config.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event not found")
		return
	}

	// Update field yang diizinkan
	event.Title = input.Title
	event.Description = input.Description
	event.Location = input.Location
	event.LocationPoint = input.LocationPoint
	event.EventType = input.EventType
	event.IsPaid = input.IsPaid
	event.Fee = input.Fee
	event.Date = input.Date
	event.Time = input.Time

	// tambah field lain sesuai kebutuhan

	// Simpan perubahan
	if err := config.DB.Save(&event).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update event")
		return
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, event)
}

func GetEvents(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	eventType := c.DefaultQuery("event_type", "")
	isPaid := c.DefaultQuery("is_paid", "")
	id := c.DefaultQuery("id", "")
	search := c.DefaultQuery("search", "")
	date := c.DefaultQuery("date", "")
	isFinish := c.DefaultQuery("is_finish", "")
	paymentType := c.DefaultQuery("payment_type", "")

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

	// Filter by Vendor
	query = query.Where("vendor_id = ?", user.VendorID)

	// Filter by Event Type
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	// Filter by ID
	if id != "" {
		query = query.Where("id = ?", id)
	}

	// Filter by IsPaid
	if isPaid != "" {
		isPaidBool, _ := strconv.ParseBool(isPaid)
		query = query.Where("is_paid = ?", isPaidBool)
	}

	// Filter by IsFinish
	if isFinish != "" {
		isFinishBool, _ := strconv.ParseBool(isFinish)
		query = query.Where("is_finish = ?", isFinishBool)
	}

	// Filter by PaymentType
	if paymentType != "" {
		query = query.Where("payment_type = ?", paymentType)
	}

	// Filter by Date (format: YYYY-MM-DD)
	if date != "" {
		query = query.Where("date = ?", date)
	}

	// Filter by Search: title, description, location
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(`
			LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(location) LIKE ?
		`, like, like, like)
	}

	// Count & Paginate
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
	if err := config.DB.First(&models.Vendor{}, input.VendorID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}
	if err := config.DB.First(&models.Event{}, input.EventID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event not found")
		return
	}

	// Cek apakah EventLog sudah ada
	var existing models.EventLog
	if err := config.DB.Where("user_id = ? AND event_id = ?", input.UserID, input.EventID).First(&existing).Error; err == nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Event log already exists for this user and event")
		return
	}

	// Buat log baru
	input.Status = true // otomatis true saat pertama kali dibuat
	if err := config.DB.Create(&input).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create event log")
		return
	}

	// Update counter user langsung
	var user models.User
	if err := config.DB.First(&user, input.UserID).Error; err == nil {
		switch strings.ToLower(input.EventType) {
		case "match":
			user.Match += 1
		case "training":
			user.Training += 1
		case "program":
			user.Program += 1
		}
		config.DB.Save(&user)
	}

	response.JSONSuccess(c.Writer, true, http.StatusCreated, gin.H{
		"message":   "Event log created successfully",
		"event_log": input,
	})
}

func UpdateEventLogStatus(c *gin.Context) {
	var input struct {
		ID     uint   `json:"id"`     // ID event log
		Status bool   `json:"status"` // status baru
		Note   string `json:"note"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Cari event log
	var eventLog models.EventLog
	if err := config.DB.First(&eventLog, input.ID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Event log not found")
		return
	}

	oldStatus := eventLog.Status

	eventLog.Status = input.Status
	eventLog.Note = input.Note
	if err := config.DB.Save(&eventLog).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update log")
		return
	}

	// Hanya update counter jika status berubah
	if oldStatus != input.Status {
		var user models.User
		if err := config.DB.First(&user, eventLog.UserID).Error; err == nil {
			switch strings.ToLower(eventLog.EventType) {
			case "match":
				if input.Status {
					user.Match++
				} else if user.Match > 0 {
					user.Match--
				}
			case "training":
				if input.Status {
					user.Training++
				} else if user.Training > 0 {
					user.Training--
				}
			case "program":
				if input.Status {
					user.Program++
				} else if user.Program > 0 {
					user.Program--
				}
			}
			config.DB.Save(&user)
		}
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"message":   "Event log updated successfully",
		"event_log": eventLog,
	})
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
