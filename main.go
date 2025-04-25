package main

import (
	"ssb_api/config"
	"ssb_api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()

	r := gin.Default()
	r.Static("/uploads", "./uploads")
	routes.SetupRoutes(r)

	r.Run(":8080")
}
