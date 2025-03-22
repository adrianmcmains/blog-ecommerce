// File: api/routes/shop_routes.go
package routes

import (
	"github.com/adrianmcmains/blog-ecommerce/api/controllers"
	"github.com/adrianmcmains/blog-ecommerce/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupShopRoutes sets up the shop-related routes
func SetupShopRoutes(router *gin.RouterGroup, shopController *controllers.ShopController, adminController *controllers.AdminController) {
	// Public routes
	router.GET("/products", shopController.GetProducts)
	router.GET("/products/:slug", shopController.GetProductBySlug)
	router.GET("/categories", shopController.GetCategories)
	router.GET("/debug-products", shopController.DebugProducts) // Debug endpoint to check products

	// Cart routes (require authentication)
	cart := router.Group("/cart")
	cart.Use(middleware.AuthMiddleware())
	{
		cart.GET("", shopController.GetCart)
		cart.POST("/items", shopController.AddToCart)
		cart.PUT("/items/:id", shopController.UpdateCartItem)
		cart.DELETE("/items/:id", shopController.RemoveFromCart)
	}

	// Order routes (require authentication)
	orders := router.Group("/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.POST("", shopController.CreateOrder)
		orders.GET("", shopController.GetOrders)
		orders.GET("/:id", shopController.GetOrderById)
	}

	// Admin routes (require admin authentication)
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RoleMiddleware("admin"))
	{
		// Product management
		admin.POST("/products", shopController.CreateProduct)
		admin.PUT("/products/:id", shopController.UpdateProduct)
		admin.DELETE("/products/:id", shopController.DeleteProduct)

		// Category management
		admin.POST("/categories", shopController.CreateCategory)
		admin.PUT("/categories/:id", shopController.UpdateCategory)
		admin.DELETE("/categories/:id", shopController.DeleteCategory)

		// Order management
		admin.GET("/orders", adminController.GetAllOrders)
		admin.GET("/orders/:id", adminController.GetOrderDetails)
		admin.PUT("/orders/:id/status", adminController.UpdateOrderStatus)
		admin.GET("/statistics", adminController.GetOrderStatistics)

		// Sync products from TinaCMS
		admin.POST("/sync-products", shopController.SyncProducts)
	}

	// Payment routes
	payment := router.Group("/payments")
	{
		// Public webhook endpoint (Eversend callbacks)
		payment.POST("/webhook", shopController.PaymentController.WebhookHandler)

		// Protected payment routes (require authentication)
		protected := payment.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/initiate", shopController.PaymentController.InitiatePayment)
			protected.GET("/:id", shopController.PaymentController.GetPaymentStatus)
			protected.POST("/:id/cancel", shopController.PaymentController.CancelPayment)
		}
	}
}