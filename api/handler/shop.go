package handler

import (
    "net/http"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/adrianmcmains/blog-ecommerce/internal/service"
    "github.com/gin-gonic/gin"
)

type ShopHandler struct {
    services *service.Service
}

func NewShopHandler(services *service.Service) *ShopHandler {
    return &ShopHandler{
        services: services,
    }
}

// Product handlers
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

func (h *ShopHandler) GetProduct(c *gin.Context) {
    id := c.Param("id")
    product, err := h.services.Shop.GetProduct(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    c.JSON(http.StatusOK, product)
}

func (h *ShopHandler) ListProducts(c *gin.Context) {
    filter := models.ProductFilter{
        Page:     c.DefaultQuery("page", "1"),
        PageSize: c.DefaultQuery("pageSize", "10"),
        Category: c.Query("category"),
        MinPrice: c.Query("minPrice"),
        MaxPrice: c.Query("maxPrice"),
    }

    products, total, err := h.services.Shop.ListProducts(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "products": products,
        "total":    total,
    })
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

// Cart handlers
func (h *ShopHandler) AddToCart(c *gin.Context) {
    var input models.AddToCartInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, _ := c.Get("userID")
    input.UserID = userID.(string)

    cartItem, err := h.services.Shop.AddToCart(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, cartItem)
}

func (h *ShopHandler) GetCart(c *gin.Context) {
    userID, _ := c.Get("userID")
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

// Order handlers
func (h *ShopHandler) CreateOrder(c *gin.Context) {
    var input models.CreateOrderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, _ := c.Get("userID")
    input.UserID = userID.(string)

    order, err := h.services.Shop.CreateOrder(c.Request.Context(), input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, order)
}

func (h *ShopHandler) GetOrder(c *gin.Context) {
    id := c.Param("id")
    userID, _ := c.Get("userID")

    order, err := h.services.Shop.GetOrder(c.Request.Context(), id, userID.(string))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
        return
    }

    c.JSON(http.StatusOK, order)
}

func (h *ShopHandler) ListOrders(c *gin.Context) {
    userID, _ := c.Get("userID")
    filter := models.OrderFilter{
        UserID:   userID.(string),
        Page:     c.DefaultQuery("page", "1"),
        PageSize: c.DefaultQuery("pageSize", "10"),
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
    })
}