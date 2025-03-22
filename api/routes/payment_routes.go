package routes

import (
	"github.com/adrianmcmains/blog-ecommerce/api/controllers"
	//"github.com/adrianmcmains/blog-ecommerce/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes sets up the payment-related routes
func SetupPaymentRoutes(router *gin.RouterGroup, controller *controllers.PaymentController, authMiddleware gin.HandlerFunc) {
	// Public routes
	router.POST("/webhook", controller.WebhookHandler)
	
	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.POST("/initiate", controller.InitiatePayment)
		protected.GET("/:id", controller.GetPaymentStatus)
		protected.POST("/:id/cancel", controller.CancelPayment)
	}
}