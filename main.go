package main

import (
	"ssb_api/config"
	"ssb_api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Mengatur Gin ke mode Release
	gin.SetMode(gin.ReleaseMode)

	// Menghubungkan ke database
	config.ConnectDatabase()

	// 2️⃣ Inisialisasi Firebase
	config.InitFirebase()

	// Membuat instance gin router
	r := gin.Default()

	// Menyajikan file statis
	r.Static("/uploads", "./uploads")

	// Menyiapkan routing
	routes.SetupRoutes(r)

	// Menjalankan server pada port 8080
	r.Run(":8080")
}
