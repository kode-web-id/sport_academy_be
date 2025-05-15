// controllers/payment_controller.go
package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"ssb_api/config"
	"ssb_api/models"
	"ssb_api/models/response"
	"ssb_api/utils"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
)

func generateInvoiceNumber(userID, vendorID uint) string {
	return fmt.Sprintf("INV-%d%d%d", vendorID, userID, time.Now().UnixNano())
}
func CreatePayment(c *gin.Context) {
	var input models.PaymentRequest

	// Binding form-data ke struct PaymentRequest (tanpa file)
	if err := c.ShouldBind(&input); err != nil {
		// Log error binding
		fmt.Println("Binding Error:", err)
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Log form data untuk debugging
	fmt.Printf("Received Form Data: %+v\n", input)

	// Get UserID from JWT token

	userID := input.UserID
	// Mengonversi PaymentRequest ke Payment
	payment := models.Payment{
		UserID:   userID,
		UserName: input.UserName,
		VendorID: input.VendorID,
		EventID:  input.EventID,
		Amount:   input.Amount,
		Method:   input.Method,
		Status:   input.Status,
		Type:     input.Type,
		Date:     input.Date,
		Note:     input.Note,
		Invoice:  generateInvoiceNumber(userID, input.VendorID),
	}

	// Mengambil vendor berdasarkan VendorID
	var vendor models.Vendor
	if err := config.DB.First(&vendor, payment.VendorID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Vendor not found")
		return
	}

	// Memeriksa apakah ada file photo yang di-upload via form-data
	file, err := c.FormFile("photo")
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "No file is attached")
		return
	}

	// Print atau log nama file yang di-upload
	fmt.Println("Uploaded Photo Filename:", file.Filename)

	// Membuat struktur direktori berdasarkan vendor_id, type_payment, dan user_id
	dstDir := fmt.Sprintf("./uploads/payment/%d/%s/%d", payment.VendorID, payment.Type, userID)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create directory for payment photo")
		return
	}

	// Membuat nama file sebagai payment_user_id_event_id
	filename := fmt.Sprintf("payment_%d_%d_%d%s", userID, *payment.EventID, time.Now().Unix(), filepath.Ext(file.Filename))

	// Menentukan lokasi file tujuan
	dst := fmt.Sprintf("%s/%s", dstDir, filename)

	// Menyimpan file yang di-upload ke tujuan
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Update path foto pada record pembayaran
	payment.Photo = dst

	// Menyimpan record pembayaran ke dalam database
	if err := config.DB.Create(&payment).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create payment")
		return
	}

	// Mengembalikan respons sukses
	response.JSONSuccess(c.Writer, true, http.StatusCreated, payment)
}
func CreateBulkPaymentByEvent(c *gin.Context) {
	// Definisikan struct untuk input langsung dari body
	var input struct {
		VendorID uint    `json:"vendor_id"`
		EventID  uint    `json:"event_id"`
		Amount   float64 `json:"amount"`
		Method   string  `json:"method"`
		Status   string  `json:"status"`
		Type     string  `json:"type"`
		Date     string  `json:"date"`
		Note     string  `json:"note"`
	}

	// Validasi input request
	if err := c.ShouldBindJSON(&input); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ambil user ID dari JWT
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	userID := uint(userIDRaw.(float64))

	// Pastikan hanya pelatih yang bisa membuat bulk payment
	var creator models.User
	if err := config.DB.First(&creator, userID).Error; err != nil || creator.Role != "pelatih" {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "Only coaches can perform this action")
		return
	}

	// Ambil semua user_id dari event_logs yang memiliki status true, berdasarkan vendor dan event
	var userIDs []uint
	if err := config.DB.
		Model(&models.EventLog{}).
		Where("vendor_id = ? AND event_id = ? AND status = ?", input.VendorID, input.EventID, true). // filter status log = true
		Pluck("user_id", &userIDs).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch user list from event logs")
		return
	}

	// Buat Payment untuk setiap user
	var createdPayments []models.Payment
	for _, uid := range userIDs {
		var user models.User
		if err := config.DB.Select("name").First(&user, uid).Error; err != nil {
			continue // skip kalau user tidak ditemukan
		}

		payment := models.Payment{
			UserID:   uid,
			UserName: user.Name, // <-- tambahkan user name
			VendorID: input.VendorID,
			EventID:  &input.EventID,
			Amount:   input.Amount,
			Method:   input.Method,
			Status:   input.Status,
			Type:     input.Type,
			Date:     input.Date,
			Note:     input.Note,
			Invoice:  generateInvoiceNumber(uid, input.VendorID),
		}
		if err := config.DB.Create(&payment).Error; err == nil {
			createdPayments = append(createdPayments, payment)
		}
	}

	// Kirim respons sukses
	response.JSONSuccess(c.Writer, true, http.StatusCreated, gin.H{
		"message":  "Bulk payment created successfully",
		"count":    len(createdPayments),
		"payments": createdPayments,
	})
}

func UploadPaymentProof(c *gin.Context) {
	// Ambil payment_id dari form data
	paymentID := c.DefaultPostForm("payment_id", "")
	if paymentID == "" {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "Payment ID is required")
		return
	}

	// Ambil user ID dari JWT token
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	userID := uint(userIDRaw.(float64))

	// Ambil data payment dari database berdasarkan payment_id
	var payment models.Payment
	if err := config.DB.First(&payment, paymentID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Payment not found")
		return
	}

	// Cek apakah user yang sedang login adalah pemiliknya
	if payment.UserID != userID {
		response.JSONErrorResponse(c.Writer, false, http.StatusForbidden, "You are not allowed to upload proof for this payment")
		return
	}

	// Ambil file 'photo' dari form data
	file, err := c.FormFile("photo")
	if err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusBadRequest, "No file uploaded or invalid file")
		return
	}

	// Tentukan direktori penyimpanan file
	dstDir := fmt.Sprintf("./uploads/payment/%d/%s/%d", payment.VendorID, payment.Type, userID)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to create directory")
		return
	}

	// Generate nama file dengan ID pembayaran dan timestamp
	filename := fmt.Sprintf("proof_%d_%d%s", payment.ID, time.Now().Unix(), filepath.Ext(file.Filename))
	dst := fmt.Sprintf("%s/%s", dstDir, filename)

	// Simpan file ke direktori yang sudah ditentukan
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Update path foto di database
	payment.Photo = dst
	if err := config.DB.Save(&payment).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update payment record")
		return
	}

	// Kirim respons sukses
	response.JSONSuccess(c.Writer, true, http.StatusOK, gin.H{
		"message": "Proof uploaded successfully",
		"photo":   dst,
	})
}

func GetPaymentsUser(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.JSONErrorResponse(c.Writer, false, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	userID := uint(userIDRaw.(float64))

	var payments []models.Payment

	// Optional: pagination
	search := c.Query("search")

	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)
	offset := (page - 1) * limit

	// Filter param
	status := c.Query("status")
	startDate := c.Query("start_date") // Format: YYYY-MM-DD
	endDate := c.Query("end_date")     // Format: YYYY-MM-DD
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	query := config.DB.Where("user_id = ?", userID)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("note ILIKE ?", like, like)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}
	baseURL := utils.DotEnv("BASE_URL_F")

	// Add base URL to any file/photo URL in payments
	for i := range payments {
		// Jika foto vendor tidak kosong, hilangkan prefix './' dan gabungkan dengan base URL
		if payments[i].Photo != "" {
			payments[i].Photo = baseURL + "/" + strings.TrimPrefix(payments[i].Photo, "./")
		}

	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, payments)
}

func GetPayments(c *gin.Context) {
	// Optional: pagination
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)
	offset := (page - 1) * limit

	// Filter param
	status := c.Query("status")
	startDate := c.Query("start_date") // Format: YYYY-MM-DD
	endDate := c.Query("end_date")     // Format: YYYY-MM-DD
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	var payments []models.Payment
	query := config.DB.Model(&models.Payment{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch payments")
		return
	}
	baseURL := utils.DotEnv("BASE_URL_F")

	// Add base URL to any file/photo URL in payments
	for i := range payments {
		// Jika foto vendor tidak kosong, hilangkan prefix './' dan gabungkan dengan base URL
		if payments[i].Photo != "" {
			payments[i].Photo = baseURL + "/" + strings.TrimPrefix(payments[i].Photo, "./")
		}
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, payments)
}

func GetPaymentsByVendor(c *gin.Context) {
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

	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	search := c.Query("search")

	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	userName := c.Query("user_name")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := strings.ToUpper(c.DefaultQuery("sort_order", "DESC"))

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)
	offset := (page - 1) * limit

	var payments []models.Payment
	query := config.DB.Where("vendor_id = ?", vendorID)

	// Filtering

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("note ILIKE ?", like, like)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if userName != "" {
		query = query.Where("user_name ILIKE ?", "%"+userName+"%")
	}

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// Execute query
	if err := query.
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to fetch vendor payments")
		return
	}

	// Add full photo path
	baseURL := utils.DotEnv("BASE_URL_F")
	for i := range payments {
		if payments[i].Photo != "" {
			payments[i].Photo = baseURL + "/" + strings.TrimPrefix(payments[i].Photo, "./")
		}
	}

	response.JSONSuccess(c.Writer, true, http.StatusOK, payments)
}

func UpdatePaymentStatus(c *gin.Context) {
	var input struct {
		PaymentID uint   `json:"payment_id"`
		Status    string `json:"status"` // "pending", "success", "failed"
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

	// Only allow trainers (pelatih) to update payment status
	if user.Role != "pelatih" {
		response.JSONErrorResponse(c.Writer, false, http.StatusForbidden, "Only trainers can update payment status")
		return
	}

	// Find the payment record
	var payment models.Payment
	if err := config.DB.First(&payment, input.PaymentID).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusNotFound, "Payment not found")
		return
	}

	// Check if the payment belongs to the same vendor as the logged-in trainer
	if payment.VendorID != user.VendorID {
		response.JSONErrorResponse(c.Writer, false, http.StatusForbidden, "You can only update payments for your own vendor")
		return
	}

	// Update the payment status
	payment.Status = input.Status
	if err := config.DB.Save(&payment).Error; err != nil {
		response.JSONErrorResponse(c.Writer, false, http.StatusInternalServerError, "Failed to update payment status")
		return
	}

	// Return success response
	response.JSONSuccess(c.Writer, true, http.StatusOK, "Update payment succesfully")
}
