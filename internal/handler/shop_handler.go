package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/service"
	"github.com/gin-gonic/gin"
)

// ShopHandler handles shop related requests
type ShopHandler struct {
    services *service.Service
    logger *log.Logger
}

// NewShopHandler creates a new shop handler
func NewShopHandler(services *service.Service, logger *log.Logger) *ShopHandler {
    return &ShopHandler{
        services: services,
        logger: logger,
    }
}

// ListProducts returns a list of products
// ListProducts godoc
// @Summary List products
// @Description Get a list of products with optional filtering
// @Tags products
// @Accept json
// @Produce json
// @Param page query string false "Page number" default(1)
// @Param pageSize query string false "Page size" default(10)
// @Param category query string false "Filter by category"
// @Param minPrice query string false "Minimum price"
// @Param maxPrice query string false "Maximum price"
// @Success 200 {object} object{products=[]models.Product,total=int}
// @Router /api/shop/products [get]
func (h *ShopHandler) ListProducts(c *gin.Context) {
    filter := models.ProductFilter{
        Page:     c.DefaultQuery("page", "1"),
        PageSize: c.DefaultQuery("page_size", "10"),
        Category: c.Query("category"),
        MinPrice: c.Query("min_price"),
        MaxPrice: c.Query("max_price"),
    }

    inStock := c.Query("in_stock")
    if inStock != "" {
        inStockBool, err := strconv.ParseBool(inStock)
        if err == nil {
            filter.InStock = &inStockBool
        }
    }

    // Get products from service
    products, total, err := h.services.Shop.ListProducts(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "products": products,
        "total":    total,
        "meta": gin.H{
            "page":      filter.Page,
            "page_size": filter.PageSize,
        },
    })
}

// GetProduct returns a single product by ID
func (h *ShopHandler) GetProduct(c *gin.Context) {
    id := c.Param("id")

    product, err := h.services.Shop.GetProduct(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    c.JSON(http.StatusOK, product)
}

// CreateProduct creates a new product
// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided input
// @Tags products
// @Accept json
// @Produce json
// @Param input body models.CreateProductInput true "Product information"
// @Success 201 {object} models.Product
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Security ApiKeyAuth
// @Router /api/shop/products [post]
func (h *ShopHandler) CreateProduct(c *gin.Context) {
    var input models.CreateProductInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    product, err := h.services.Shop.CreateProduct(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func (h *ShopHandler) UpdateProduct(c *gin.Context) {
    id := c.Param("id")

    var input models.CreateProductInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    product, err := h.services.Shop.UpdateProduct(c.Request.Context(), id, input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product
func (h *ShopHandler) DeleteProduct(c *gin.Context) {
    id := c.Param("id")

    if err := h.services.Shop.DeleteProduct(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// AddToCart adds a product to the user's cart
func (h *ShopHandler) AddToCart(c *gin.Context) {
    var input models.AddToCartInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }
    input.UserID = userID.(string)

    cartItem, err := h.services.Shop.AddToCart(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, cartItem)
}

// GetCart returns the user's cart
func (h *ShopHandler) GetCart(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }

    cart, err := h.services.Shop.GetCart(c.Request.Context(), userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, cart)
}

// UpdateCartItem updates a cart item's quantity
func (h *ShopHandler) UpdateCartItem(c *gin.Context) {
    id := c.Param("id")

    var input struct {
        Quantity int `json:"quantity" binding:"required,gt=0"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    cartItem, err := h.services.Shop.UpdateCartItem(c.Request.Context(), id, input.Quantity)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, cartItem)
}

// RemoveFromCart removes an item from the cart
func (h *ShopHandler) RemoveFromCart(c *gin.Context) {
    id := c.Param("id")

    if err := h.services.Shop.RemoveFromCart(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// CreateOrder creates a new order from cart items
func (h *ShopHandler) CreateOrder(c *gin.Context) {
    var input models.CreateOrderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }
    input.UserID = userID.(string)

    order, err := h.services.Shop.CreateOrder(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, order)
}

// GetOrder returns a single order by ID
func (h *ShopHandler) GetOrder(c *gin.Context) {
    id := c.Param("id")

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }

    order, err := h.services.Shop.GetOrder(c.Request.Context(), id, userID.(string))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
        return
    }

    c.JSON(http.StatusOK, order)
}

// ListOrders returns a list of orders for the authenticated user
func (h *ShopHandler) ListOrders(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }

    filter := models.OrderFilter{
        UserID:   userID.(string),
        Page:     c.DefaultQuery("page", "1"),
        PageSize: c.DefaultQuery("page_size", "10"),
        Status:   c.Query("status"),
    }

    orders, total, err := h.services.Shop.ListOrders(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "orders": orders,
        "total":  total,
        "meta": gin.H{
            "page":      filter.Page,
            "page_size": filter.PageSize,
        },
    })
}