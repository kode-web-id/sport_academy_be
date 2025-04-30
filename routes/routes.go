package routes

import (
	"ssb_api/controllers"
	"ssb_api/controllers/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Public routes
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.Register)
		api.GET("/vendor", controllers.GetVendors)
		api.POST("/vendor/create", controllers.CreateVendor)
		api.DELETE("/vendor/:id", controllers.DeleteVendorByID)

		// Protected routes with JWT middleware
		protected := api.Group("", middleware.JWTAuthMiddleware())
		{
			// Users
			protected.GET("/user/profile", controllers.GetUserFromToken)
			protected.GET("/users", controllers.GetAllUsers)
			protected.GET("/users/vendor", controllers.GetUsersByVendor)
			protected.GET("/users/search", controllers.SearchUsers)
			protected.PUT("/user/foto", controllers.UpdateUserPhoto)
			protected.PUT("/vendor/foto", controllers.UpdateVendorPhoto)
			protected.PUT("/vendor/bank", controllers.UpdateVendorBank)
			protected.PUT("/user/update", controllers.UpdateUser)

			// Payments
			protected.GET("/payments/user", controllers.GetPaymentsUser)
			protected.GET("/payments", controllers.GetPayments)
			protected.POST("/payment/create", controllers.CreatePayment)
			protected.GET("/payments/vendor", controllers.GetPaymentsByVendor)
			protected.PUT("/payment/status", controllers.UpdatePaymentStatus)
			protected.POST("/payment/bulk", controllers.CreateBulkPaymentByEvent)
			protected.PUT("/payment/proof", controllers.UploadPaymentProof)

			// Trainings
			protected.GET("/trainings", controllers.GetTrainings)
			protected.POST("/training/create", controllers.CreateTraining)
			protected.GET("/trainings/vendor", controllers.GetTrainingsByVendor)

			// Match
			protected.GET("/matches", controllers.GetMatchs)
			protected.POST("/match/create", controllers.CreateMatch)
			protected.GET("/matches/vendor", controllers.GetMatchsByVendor)

			// Challenges
			protected.GET("/challenges", controllers.GetChallenges)
			protected.POST("/challenge/create", controllers.CreateChallenge)
			protected.GET("/challenge-logs", controllers.GetChallengeLogs)
			protected.POST("/challenge-log/create", controllers.CreateChallengeLog)
			protected.GET("/challenges/vendor", controllers.GetChallengesByVendor)

			// Events
			protected.GET("/events", controllers.GetEvents)
			protected.POST("/event/create", controllers.CreateEvent)
			protected.GET("/event-logs", controllers.GetEventLogs)
			protected.POST("/event-log/create", controllers.CreateEventLog)
			protected.GET("/event-logs/user", controllers.GetEventLogsByUser)
			protected.PUT("/event-log/status", controllers.UpdateEventLogStatus)
			protected.PUT("/events/finish", controllers.UpdateEventFinishStatus)

		}

	}
}
