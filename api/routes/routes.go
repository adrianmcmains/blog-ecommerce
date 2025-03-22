package routes

import (
	"database/sql"
	"log"

	"github.com/adrianmcmains/blog-ecommerce/api/controllers"
	"github.com/adrianmcmains/blog-ecommerce/api/handler"
	"github.com/adrianmcmains/blog-ecommerce/api/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupRouter(r *gin.Engine, handler *handler.Handler, db *sql.DB) {
    // Convert sql.DB to gorm.DB
    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: db,
    }), &gorm.Config{})
    
    if err != nil {
        log.Fatalf("Failed to convert sql.DB to gorm.DB: %v", err)
    }
    
    // Initialize controllers
    authController := controllers.NewAuthController(db)
    blogController := controllers.NewBlogController(gormDB) // Use gormDB here
    shopController := controllers.NewShopController(db)
    adminController := controllers.NewAdminController(db)
    
    // Initialize payment controller
    paymentController, err := controllers.NewPaymentController(db)
    if err != nil {
        log.Printf("Warning: Failed to initialize payment controller: %v", err)
    } else {
        shopController.PaymentController = paymentController
    }

    // Middleware (global middleware)
    r.Use(middleware.CORSMiddleware())
    r.Use(middleware.SecurityHeaders())

    // API routes
    api := r.Group("/api")
    {
        // Auth routes
        auth := api.Group("/auth")
        {
            // Handler-based routes
            auth.POST("/register", handler.Auth.Register)
            auth.POST("/login", handler.Auth.Login)
            auth.GET("/me", middleware.AuthMiddleware(), handler.Auth.GetCurrentUser)
            
            // Controller-based routes
            auth.POST("/change-password", middleware.AuthMiddleware(), authController.ChangePassword)
            auth.POST("/logout", authController.Logout)
            
            // Comment these out until implemented
            // auth.POST("/refresh-token", authController.RefreshToken)
            // auth.POST("/forgot-password", authController.ForgotPassword)
            // auth.POST("/reset-password", authController.ResetPassword)
        }

        // Blog routes
        blog := api.Group("/blog")
        {
            // Public routes
            blog.GET("/posts", handler.Blog.ListPosts)
            blog.GET("/posts/:id", handler.Blog.GetPost)
            blog.GET("/categories", blogController.GetCategories)
            blog.GET("/tags", blogController.GetTags)
            
            // Comment this out until implemented
            // blog.GET("/authors", blogController.GetAuthors)
            
            // Protected routes - require authentication and proper role
            blogAdmin := blog.Group("/")
            blogAdmin.Use(middleware.AuthMiddleware())
            {
                // Admin/contributor routes
                adminContributor := blogAdmin.Group("/")
                adminContributor.Use(middleware.RequireRole("admin", "contributor"))
                {
                    adminContributor.POST("/posts", handler.Blog.CreatePost)
                    adminContributor.PUT("/posts/:id", handler.Blog.UpdatePost)
                    
                    // Comment these out until implemented
                    // adminContributor.POST("/draft", blogController.SaveDraft)
                    // adminContributor.PUT("/draft/:id", blogController.UpdateDraft)
                }
                
                // Admin-only routes
                adminOnly := blogAdmin.Group("/")
                adminOnly.Use(middleware.RequireRole("admin"))
                {
                    adminOnly.DELETE("/posts/:id", handler.Blog.DeletePost)
                    adminOnly.POST("/categories", blogController.CreateCategory)
                    adminOnly.PUT("/categories/:id", blogController.UpdateCategory)
                    adminOnly.DELETE("/categories/:id", blogController.DeleteCategory)
                    adminOnly.POST("/tags", blogController.CreateTag)
                    adminOnly.PUT("/tags/:id", blogController.UpdateTag)
                    adminOnly.DELETE("/tags/:id", blogController.DeleteTag)
                    
                    // Comment this out until implemented
                    // adminOnly.POST("/sync", blogController.SyncContent)
                }
                
                // Comment routes
                comments := blogAdmin.Group("/comments")
                {
                    comments.POST("", blogController.AddComment)
                    
                    // Comment these out until implemented
                    // comments.PUT("/:id", blogController.UpdateOwnComment)
                    // comments.DELETE("/:id", blogController.DeleteOwnComment)
                }
            }
        }

        // Shop routes
        shop := api.Group("/shop")
        {
            // Public product routes
            shop.GET("/products", handler.Shop.ListProducts)
            shop.GET("/products/:id", handler.Shop.GetProduct)
            shop.GET("/products/slug/:slug", shopController.GetProductBySlug)
            shop.GET("/categories", shopController.GetCategories)
            shop.GET("/featured", shopController.GetFeaturedProducts)
            shop.GET("/new-arrivals", shopController.GetNewArrivals)
            shop.GET("/best-sellers", shopController.GetBestSellers)
            shop.GET("/search", shopController.SearchProducts)
            shop.GET("/debug-products", shopController.DebugProducts) // Debug endpoint

            // Admin product routes
            productAdmin := shop.Group("/admin")
            productAdmin.Use(middleware.AuthMiddleware())
            productAdmin.Use(middleware.RequireRole("admin"))
            {
                productAdmin.POST("/products", handler.Shop.CreateProduct)
                productAdmin.PUT("/products/:id", handler.Shop.UpdateProduct)
                productAdmin.DELETE("/products/:id", handler.Shop.DeleteProduct)
                productAdmin.POST("/categories", shopController.CreateCategory)
                productAdmin.PUT("/categories/:id", shopController.UpdateCategory)
                productAdmin.DELETE("/categories/:id", shopController.DeleteCategory)
                productAdmin.POST("/sync-products", shopController.SyncProducts)
                productAdmin.GET("/inventory", shopController.GetInventoryReport)
                productAdmin.PUT("/inventory/update-stock", shopController.UpdateStockLevels)
                
                // Order management
                productAdmin.GET("/orders", adminController.GetAllOrders)
                productAdmin.GET("/orders/:id", adminController.GetOrderDetails)
                productAdmin.PUT("/orders/:id/status", adminController.UpdateOrderStatus)
                productAdmin.GET("/statistics", adminController.GetOrderStatistics)
                
                // Comment this out until implemented
                // productAdmin.GET("/dashboard", adminController.GetDashboardData)
            }

            // Cart routes (require authentication)
            cart := shop.Group("/cart")
            cart.Use(middleware.AuthMiddleware())
            {
                cart.GET("", handler.Shop.GetCart)
                cart.POST("/items", handler.Shop.AddToCart)
                cart.PUT("/items/:id", handler.Shop.UpdateCartItem)
                cart.DELETE("/items/:id", handler.Shop.RemoveFromCart)
                cart.DELETE("", shopController.ClearCart)
                cart.GET("/count", shopController.GetCartItemCount)
            }

            // Order routes (require authentication)
            orders := shop.Group("/orders")
            orders.Use(middleware.AuthMiddleware())
            {
                orders.POST("", handler.Shop.CreateOrder)
                orders.GET("", handler.Shop.ListOrders)
                orders.GET("/:id", handler.Shop.GetOrder)
                orders.GET("/:id/tracking", shopController.GetOrderTracking)
                orders.POST("/:id/cancel", shopController.CancelOrder)
            }
            
            // Wishlist routes (require authentication)
            wishlist := shop.Group("/wishlist")
            wishlist.Use(middleware.AuthMiddleware())
            {
                wishlist.GET("", shopController.GetWishlist)
                wishlist.POST("", shopController.AddToWishlist)
                wishlist.DELETE("/:id", shopController.RemoveFromWishlist)
                wishlist.POST("/:id/move-to-cart", shopController.MoveWishlistItemToCart)
            }
            
            // Reviews routes
            reviews := shop.Group("/reviews")
            {
                reviews.GET("/product/:id", shopController.GetProductReviews)
                
                // Protected review routes
                authReviews := reviews.Group("/")
                authReviews.Use(middleware.AuthMiddleware())
                {
                    authReviews.POST("", shopController.AddReview)
                    authReviews.PUT("/:id", shopController.UpdateOwnReview)
                    authReviews.DELETE("/:id", shopController.DeleteOwnReview)
                }
            }
        }
        
        // Payment routes
        payments := api.Group("/payments")
        {
            // Public webhook for payment provider callbacks
            payments.POST("/webhook", shopController.PaymentController.WebhookHandler)
            
            // Protected payment routes
            authPayments := payments.Group("/")
            authPayments.Use(middleware.AuthMiddleware())
            {
                authPayments.POST("/initiate", shopController.PaymentController.InitiatePayment)
                authPayments.GET("/:id", shopController.PaymentController.GetPaymentStatus)
                authPayments.POST("/:id/cancel", shopController.PaymentController.CancelPayment)
                authPayments.GET("/methods", shopController.PaymentController.GetPaymentMethods)
            }
        }
        
        // Content management routes (for TinaCMS integration)
        cms := api.Group("/cms")
        cms.Use(middleware.AuthMiddleware())
        cms.Use(middleware.RequireRole("admin", "contributor"))
        {
            cms.POST("/sync", shopController.SyncContentFromTina)
        }
        
        // User profile routes
        profile := api.Group("/profile")
        profile.Use(middleware.AuthMiddleware())
        {
            // Comment these out until implemented
            // profile.GET("", handler.Auth.GetProfile)
            // profile.PUT("", handler.Auth.UpdateProfile)
            
            profile.GET("/addresses", shopController.GetUserAddresses)
            profile.POST("/addresses", shopController.AddUserAddress)
            profile.PUT("/addresses/:id", shopController.UpdateUserAddress)
            profile.DELETE("/addresses/:id", shopController.DeleteUserAddress)
        }
    }
    
    // Admin dashboard API
    admin := r.Group("/api/admin")
    admin.Use(middleware.AuthMiddleware())
    admin.Use(middleware.RequireRole("admin"))
    {
        // Comment these out until implemented
        // admin.GET("/users", adminController.GetAllUsers)
        // admin.GET("/users/:id", adminController.GetUser)
        // admin.PUT("/users/:id", adminController.UpdateUser)
        // admin.DELETE("/users/:id", adminController.DeleteUser)
        // admin.PUT("/users/:id/role", adminController.ChangeUserRole)
        
        // admin.GET("/logs", adminController.GetSystemLogs)
        // admin.GET("/statistics/site", adminController.GetSiteStatistics)
        // admin.GET("/settings", adminController.GetSystemSettings)
        // admin.PUT("/settings", adminController.UpdateSystemSettings)
        // admin.POST("/db/backup", adminController.BackupDatabase)
        // admin.POST("/db/restore", adminController.RestoreDatabase)
        // admin.GET("/health", adminController.GetSystemHealth)
    }
    
    // Testing/debug routes - only available in development
    if gin.Mode() != gin.ReleaseMode {
        debug := r.Group("/api/debug")
        {
            debug.GET("/routes", func(c *gin.Context) {
                routes := r.Routes()
                c.JSON(200, routes)
            })
            debug.GET("/config", middleware.AuthMiddleware(), middleware.RequireRole("admin"), func(c *gin.Context) {
                // Return non-sensitive configuration info for debugging
                c.JSON(200, gin.H{
                    "environment": gin.Mode(),
                    "version": "1.0.0",
                    "db_connected": db != nil,
                })
            })
        }
    }
}