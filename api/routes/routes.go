package routes

import (
	"github.com/adrianmcmains/blog-ecommerce/api/handler"
	"github.com/adrianmcmains/blog-ecommerce/api/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, handler *handler.Handler, jwtSecret string) {
    // Middleware
    r.Use(middleware.CORSMiddleware())

    // API routes
    api := r.Group("/api")
    {
        // Auth routes
        auth := api.Group("/auth")
        {
            auth.POST("/register", handler.Auth.Register)
            auth.POST("/login", handler.Auth.Login)
            auth.GET("/me", middleware.AuthMiddleware(jwtSecret), handler.Auth.GetCurrentUser)
        }

        // Blog routes
        blog := api.Group("/blog")
        {
            blog.GET("/posts", handler.Blog.ListPosts)
            blog.GET("/posts/:id", handler.Blog.GetPost)
            blog.POST("/posts", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin", "contributor"), handler.Blog.CreatePost)
            blog.PUT("/posts/:id", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin", "contributor"), handler.Blog.UpdatePost)
            blog.DELETE("/posts/:id", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin"), handler.Blog.DeletePost)
        }

        // Shop routes
        shop := api.Group("/shop")
        {
            // Product routes
            shop.GET("/products", handler.Shop.ListProducts)
            shop.GET("/products/:id", handler.Shop.GetProduct)
            shop.POST("/products", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin"), handler.Shop.CreateProduct)
            shop.PUT("/products/:id", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin"), handler.Shop.UpdateProduct)
            shop.DELETE("/products/:id", middleware.AuthMiddleware(jwtSecret), middleware.RequireRole("admin"), handler.Shop.DeleteProduct)

            // Cart routes
            cart := shop.Group("/cart", middleware.AuthMiddleware(jwtSecret))
            {
                cart.GET("", handler.Shop.GetCart)
                cart.POST("/items", handler.Shop.AddToCart)
                cart.PUT("/items/:id", handler.Shop.UpdateCartItem)
                cart.DELETE("/items/:id", handler.Shop.RemoveFromCart)  
            }

            // Order routes
            orders := shop.Group("/orders", middleware.AuthMiddleware(jwtSecret))
            {
                orders.POST("", handler.Shop.CreateOrder)
                orders.GET("", handler.Shop.ListOrders)
                orders.GET("/:id", handler.Shop.GetOrder)
            }
        }
    }
}