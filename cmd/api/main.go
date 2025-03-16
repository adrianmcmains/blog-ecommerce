// @title Blog & E-commerce API
// @version 1.0
// @description API for managing blog posts and e-commerce functionality
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adrianmcmains/blog-ecommerce/internal/handler"
	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/repository"
	"github.com/adrianmcmains/blog-ecommerce/internal/service"
	"github.com/adrianmcmains/blog-ecommerce/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var services *service.Service
var logger *log.Logger

func init() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Printf("No .env file found")
    }
    
    // Initialize logger
    logger = log.New(os.Stdout, "[BLOG-ECOMMERCE] ", log.LstdFlags)
}

func main() {
    // Initialize database
    initDB()
    
    // Initialize repositories
    repos := repository.NewRepository(db)
    
    // Create token repository
    tokenRepo := repository.NewTokenRepository(db)
    
    // Initialize services
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "your-default-secret-key" // Not secure for production
    }
    jwtTTL := 24 * time.Hour
    refreshTTL := 7 * 24 * time.Hour
    
    services = service.NewService(repos, tokenRepo, jwtSecret, jwtTTL, refreshTTL)

    // Set up router
    r := gin.Default()

    // CORS middleware
    r.Use(corsMiddleware())

    // Initialize handlers
    handlers := handler.NewHandler(services, logger)

    // API Routes
    api := r.Group("/api")
    {
        // Auth routes
        auth := api.Group("/auth")
        {
            auth.POST("/register", registerHandler)
            auth.POST("/login", loginHandler)
            auth.GET("/me", authMiddleware(), getCurrentUser)
        }

        // Blog routes
        blog := api.Group("/blog")
        {
            blog.GET("/posts", handlers.Blog.GetAllPosts)
            blog.GET("/posts/:id", handlers.Blog.GetPostByID)
            blog.POST("/posts", authMiddleware(), handlers.Blog.CreatePost)
            blog.PUT("/posts/:id", authMiddleware(), handlers.Blog.UpdatePost)
            blog.DELETE("/posts/:id", authMiddleware(), handlers.Blog.DeletePost)
        }

        // Shop routes
        shop := api.Group("/shop")
        {
            shop.GET("/products", handlers.Shop.ListProducts)
            shop.GET("/products/:id", getProduct)
            shop.POST("/products", authMiddleware(), createProduct)
            shop.PUT("/products/:id", authMiddleware(), updateProduct)
            shop.DELETE("/products/:id", authMiddleware(), deleteProduct)

            // Cart and Order routes
            shop.POST("/cart", authMiddleware(), addToCart)
            shop.GET("/cart", authMiddleware(), getCart)
            shop.DELETE("/cart/:id", authMiddleware(), removeFromCart)
            shop.POST("/orders", authMiddleware(), createOrder)
            shop.GET("/orders", authMiddleware(), listOrders)
        }
    }

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    logger.Printf("Server starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            c.Abort()
            return
        }

        // Extract the token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        // Get JWT secret from environment
        jwtSecret := os.Getenv("JWT_SECRET")
        if jwtSecret == "" {
            jwtSecret = "your-default-secret-key" // Not secure for production
        }

        // Parse and validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Extract claims
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Set user info in context
        c.Set("userID", claims["user_id"])
        c.Set("userRole", claims["role"])
        c.Next()
    }
}

func initDB() {
    var err error
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=localhost user=postgres password=postgres dbname=blog_ecommerce port=5432 sslmode=disable"
    }
    
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto migrate schemas
    err = db.AutoMigrate(
        &models.User{},
        &models.Post{},
        &models.Category{},
        &models.Tag{},
        &models.Product{},
        &models.ProductVariant{},
        &models.CartItem{},
        &models.Order{},
        &models.OrderItem{},
    )
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
}

// Auth Handlers
func registerHandler(c *gin.Context) {
    var input models.RegisterInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check if user already exists
    var existingUser models.User
    if result := db.Where("email = ?", input.Email).First(&existingUser); result.Error == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
        return
    }

    // Create user
    user := models.User{
        Email:     input.Email,
        FirstName: input.FirstName,
        LastName:  input.LastName,
        Role:      input.Role,
    }

    if err := user.SetPassword(input.Password); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    // Generate JWT token
    token, err := generateToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "user":  user,
        "token": token,
    })
}

func loginHandler(c *gin.Context) {
    var input models.LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    if !user.CheckPassword(input.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }

    token, err := generateToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}

func getCurrentUser(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }

    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, user)
}

func generateToken(user models.User) (string, error) {
    // Get JWT secret from environment
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "your-default-secret-key" // Not secure for production
    }

    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "role":    user.Role,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })

    // Sign and get the complete encoded token as a string
    tokenString, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

// Shop Handlers
func getProduct(c *gin.Context) {
    id := c.Param("id")
    
    var product models.Product
    if err := db.Preload("Categories").Preload("Variants").First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
    
    c.JSON(http.StatusOK, product)
}

func createProduct(c *gin.Context) {
    var input models.CreateProductInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    product := models.Product{
        Name:          input.Name,
        Slug:          util.GenerateSlug(input.Name),
        Description:   input.Description,
        Price:         input.Price,
        StockQuantity: input.StockQuantity,
        Image:         input.Image,
        Status:        "active",
    }
    
    // Handle categories
    if len(input.Categories) > 0 {
        for _, categoryName := range input.Categories {
            var category models.Category
            
            // Try to find existing category
            result := db.Where("name = ?", categoryName).First(&category)
            if result.Error != nil {
                // Create new category
                category = models.Category{
                    Name: categoryName,
                    Slug: util.GenerateSlug(categoryName),
                }
                db.Create(&category)
            }
            
            product.Categories = append(product.Categories, category)
        }
    }
    
    // Handle variants
    if len(input.Variants) > 0 {
        for _, variantInput := range input.Variants {
            variant := models.ProductVariant{
                Name:          variantInput.Name,
                Price:         variantInput.Price,
                StockQuantity: variantInput.StockQuantity,
                SKU:           variantInput.SKU,
            }
            
            product.Variants = append(product.Variants, variant)
        }
    }
    
    if err := db.Create(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
        return
    }
    
    c.JSON(http.StatusCreated, product)
}

func updateProduct(c *gin.Context) {
    id := c.Param("id")
    
    var product models.Product
    if err := db.First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
    
    var input models.CreateProductInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Update fields
    product.Name = input.Name
    product.Slug = util.GenerateSlug(input.Name)
    product.Description = input.Description
    product.Price = input.Price
    product.StockQuantity = input.StockQuantity
    product.Image = input.Image
    
    if err := db.Save(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
        return
    }
    
    c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
    id := c.Param("id")
    
    if err := db.Delete(&models.Product{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func addToCart(c *gin.Context) {
    var input models.AddToCartInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID, _ := c.Get("userID")
    
    // Check if product exists
    var product models.Product
    if err := db.First(&product, input.ProductID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
    
    // Check stock
    if product.StockQuantity < input.Quantity {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock available"})
        return
    }
    
    // Create cart item
    userIDStr, ok := userID.(string)
    if !ok {
        userIDStr = userID.(uuid.UUID).String()
    }
    
    userUUID, err := uuid.Parse(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    productUUID, err := uuid.Parse(input.ProductID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    
    cartItem := models.CartItem{
        UserID:    userUUID,
        ProductID: productUUID,
        Quantity:  input.Quantity,
    }
    
    // Add variant if specified
    if input.VariantID != "" {
        variantUUID, err := uuid.Parse(input.VariantID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
            return
        }
        cartItem.VariantID = &variantUUID
    }
    
    if err := db.Create(&cartItem).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
        return
    }
    
    c.JSON(http.StatusCreated, cartItem)
}

func getCart(c *gin.Context) {
    userID, _ := c.Get("userID")
    
    userIDStr, ok := userID.(string)
    if !ok {
        userIDStr = userID.(uuid.UUID).String()
    }
    
    var cartItems []models.CartItem
    if err := db.Where("user_id = ?", userIDStr).
              Preload("Product").
              Preload("Variant").
              Find(&cartItems).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
        return
    }
    
    c.JSON(http.StatusOK, cartItems)
}

func removeFromCart(c *gin.Context) {
    id := c.Param("id")
    
    if err := db.Delete(&models.CartItem{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

func createOrder(c *gin.Context) {
    var input models.CreateOrderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID, _ := c.Get("userID")
    userIDStr, ok := userID.(string)
    if !ok {
        userIDStr = userID.(uuid.UUID).String()
    }
    
    userUUID, err := uuid.Parse(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    // Get cart items
    var cartItems []models.CartItem
    var totalAmount float64
    
    for _, cartID := range input.CartIDs {
        var cartItem models.CartItem
        if err := db.Preload("Product").Preload("Variant").First(&cartItem, cartID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
            return
        }
        
        cartItems = append(cartItems, cartItem)
        
        // Calculate price
        price := cartItem.Product.Price
        if cartItem.Variant != nil {
            price = cartItem.Variant.Price
        }
        
        totalAmount += price * float64(cartItem.Quantity)
    }
    
    // Create order
    order := models.Order{
        UserID:      userUUID,
        Status:      "pending",
        TotalAmount: totalAmount,
        Address: models.Address{
            Street:     input.Address.Street,
            City:       input.Address.City,
            State:      input.Address.State,
            Country:    input.Address.Country,
            PostalCode: input.Address.PostalCode,
        },
    }
    
    // Add order items
    for _, cartItem := range cartItems {
        price := cartItem.Product.Price
        if cartItem.Variant != nil {
            price = cartItem.Variant.Price
        }
        
        orderItem := models.OrderItem{
            ProductID:   cartItem.ProductID,
            Quantity:    cartItem.Quantity,
            PriceAtTime: price,
        }
        
        if cartItem.VariantID != nil {
            orderItem.VariantID = cartItem.VariantID
        }
        
        order.Items = append(order.Items, orderItem)
    }
    
    // Create transaction
    tx := db.Begin()
    
    // Save order
    if err := tx.Create(&order).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
        return
    }
    
    // Clear cart items
    for _, cartItem := range cartItems {
        if err := tx.Delete(&cartItem).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
            return
        }
        
        // Update product stock
        if err := tx.Model(&models.Product{}).
                  Where("id = ?", cartItem.ProductID).
                  Update("stock_quantity", gorm.Expr("stock_quantity - ?", cartItem.Quantity)).
                  Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
            return
        }
    }
    
    tx.Commit()
    
    c.JSON(http.StatusCreated, order)
}

func listOrders(c *gin.Context) {
    userID, _ := c.Get("userID")
    
    userIDStr, ok := userID.(string)
    if !ok {
        userIDStr = userID.(uuid.UUID).String()
    }
    
    var orders []models.Order
    if err := db.Where("user_id = ?", userIDStr).
              Preload("Items").
              Preload("Items.Product").
              Preload("Items.Variant").
              Find(&orders).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
        return
    }
    
    c.JSON(http.StatusOK, orders)
}