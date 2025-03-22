// api/controllers/order_controller.go
package controllers

import (
	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateOrder creates a new order from the cart
func (c *ShopController) CreateOrder(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req struct {
		ShippingAddress string `json:"shippingAddress" binding:"required"`
		BillingAddress  string `json:"billingAddress"`
		PaymentMethod   string `json:"paymentMethod" binding:"required"`
		Notes           string `json:"notes"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If billing address is not provided, use shipping address
	if req.BillingAddress == "" {
		req.BillingAddress = req.ShippingAddress
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get user's cart
	var cartID int
	err = tx.QueryRow(`
		SELECT id
		FROM carts
		WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Your cart is empty"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		}
		return
	}

	// Get cart items
	rows, err := tx.Query(`
		SELECT id, product_id, name, price, quantity
		FROM cart_items
		WHERE cart_id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart items"})
		return
	}
	defer rows.Close()

	// Verify cart has items
	if !rows.Next() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Your cart is empty"})
		return
	}
	rows.Close()

	// Calculate total order amount
	var totalAmount float64
	err = tx.QueryRow(`
		SELECT SUM(price * quantity)
		FROM cart_items
		WHERE cart_id = $1
	`, cartID).Scan(&totalAmount)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total"})
		return
	}

	// Create order
	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (
			user_id, total_amount, status, shipping_address, billing_address,
			payment_method, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, userID, totalAmount, models.OrderStatusPending, req.ShippingAddress,
		req.BillingAddress, req.PaymentMethod, req.Notes).Scan(&orderID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Move cart items to order items
	rows, err = tx.Query(`
		SELECT product_id, name, price, quantity
		FROM cart_items
		WHERE cart_id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart items"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var productID int
		var productName string
		var price float64
		var quantity int

		err := rows.Scan(&productID, &productName, &price, &quantity)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process cart items"})
			return
		}

		// Check stock availability
		var inStock int
		err = tx.QueryRow(`
			SELECT stock
			FROM products
			WHERE id = $1
		`, productID).Scan(&inStock)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check product stock"})
			return
		}

		if inStock < quantity {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Not enough stock available for product: " + productName,
			})
			return
		}

		// Add to order items
		_, err = tx.Exec(`
			INSERT INTO order_items (
				order_id, product_id, product_name, quantity, unit_price, total_price, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, orderID, productID, productName, quantity, price, price*float64(quantity))

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order items"})
			return
		}

		// Update product stock
		_, err = tx.Exec(`
			UPDATE products
			SET stock = stock - $1
			WHERE id = $2
		`, quantity, productID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	// Clear user's cart
	_, err = tx.Exec(`
		DELETE FROM cart_items
		WHERE cart_id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get created order
	order, err := c.getOrderByID(orderID, userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get created order"})
		return
	}

	ctx.JSON(http.StatusCreated, order)
}

// GetOrders gets all orders for the current user
func (c *ShopController) GetOrders(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	status := ctx.Query("status")

	// Build query
	query := `
		SELECT id, user_id, total_amount, status, transaction_id,
			shipping_address, billing_address, payment_method, notes,
			created_at, updated_at
		FROM orders
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	// Execute query
	rows, err := c.db.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	// Process orders
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.TransactionID,
			&order.ShippingAddress, &order.BillingAddress, &order.PaymentMethod, &order.Notes,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process orders"})
			return
		}
		orders = append(orders, order)
	}

	// Count total orders
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	countArgs := []interface{}{userID}

	if status != "" {
		countQuery += " AND status = $2"
		countArgs = append(countArgs, status)
	}

	err = c.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count orders"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"total":  totalCount,
		"limit":  limit,
		"offset": offset,
	})
}

// GetOrderById gets a specific order by ID
func (c *ShopController) GetOrderById(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get order ID from URL
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Get order
	order, err := c.getOrderByID(orderID, userID.(int))
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// Helper to get an order by ID with items
func (c *ShopController) getOrderByID(orderID, userID int) (models.Order, error) {
	var order models.Order

	// Get order details
	err := c.db.QueryRow(`
		SELECT id, user_id, total_amount, status, transaction_id,
			shipping_address, billing_address, payment_method, notes,
			created_at, updated_at
		FROM orders
		WHERE id = $1 AND user_id = $2
	`, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.TransactionID,
		&order.ShippingAddress, &order.BillingAddress, &order.PaymentMethod, &order.Notes,
		&order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return order, err
	}

	// Get order items
	rows, err := c.db.Query(`
		SELECT id, product_id, product_name, quantity, unit_price, total_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return order, err
	}
	defer rows.Close()

	// Process order items
	order.Items = []models.OrderItem{}
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID, &item.ProductID, &item.ProductName, &item.Quantity,
			&item.UnitPrice, &item.TotalPrice, &item.CreatedAt,
		)
		if err != nil {
			return order, err
		}

		item.OrderID = orderID
		order.Items = append(order.Items, item)
	}

	return order, nil
}