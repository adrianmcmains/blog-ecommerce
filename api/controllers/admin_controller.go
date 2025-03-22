// api/controllers/admin_controller.go
package controllers

import (
	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminController handles admin-related requests
type AdminController struct {
	db *sql.DB
}

// NewAdminController creates a new admin controller
func NewAdminController(db *sql.DB) *AdminController {
	return &AdminController{db: db}
}

// GetAllOrders gets all orders (admin only)
func (c *AdminController) GetAllOrders(ctx *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	status := ctx.Query("status")
	userID := ctx.Query("userId")
	sortBy := ctx.DefaultQuery("sort", "created_at")
	sortOrder := ctx.DefaultQuery("order", "desc")

	// Build query
	query := `
		SELECT o.id, o.user_id, o.total_amount, o.status, o.transaction_id,
			o.shipping_address, o.billing_address, o.payment_method, o.notes,
			o.created_at, o.updated_at,
			u.name as user_name, u.email as user_email
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if status != "" {
		query += " AND o.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	if userID != "" {
		query += " AND o.user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
		argCount++
	}

	// Add sorting
	validSortColumns := map[string]bool{
		"id": true, "user_id": true, "total_amount": true, "status": true, "created_at": true, "updated_at": true,
	}
	if validSortColumns[sortBy] {
		query += " ORDER BY o." + sortBy
	} else {
		query += " ORDER BY o.created_at"
	}

	if sortOrder == "asc" {
		query += " ASC"
	} else {
		query += " DESC"
	}

	// Add pagination
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := c.db.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	// Process orders
	type OrderWithUserInfo struct {
		models.Order
		UserName  string `json:"userName"`
		UserEmail string `json:"userEmail"`
	}

	var orders []OrderWithUserInfo
	for rows.Next() {
		var order OrderWithUserInfo
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.TransactionID,
			&order.ShippingAddress, &order.BillingAddress, &order.PaymentMethod, &order.Notes,
			&order.CreatedAt, &order.UpdatedAt,
			&order.UserName, &order.UserEmail,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process orders"})
			return
		}
		orders = append(orders, order)
	}

	// Count total orders
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM orders o
		WHERE 1=1
	`
	countArgs := []interface{}{}
	argCount = 1

	if status != "" {
		countQuery += " AND o.status = $" + strconv.Itoa(argCount)
		countArgs = append(countArgs, status)
		argCount++
	}

	if userID != "" {
		countQuery += " AND o.user_id = $" + strconv.Itoa(argCount)
		countArgs = append(countArgs, userID)
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

// GetOrderDetails gets details of a specific order (admin only)
func (c *AdminController) GetOrderDetails(ctx *gin.Context) {
	// Get order ID from URL
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Get order details with user info
	type OrderWithUserInfo struct {
		models.Order
		UserName  string `json:"userName"`
		UserEmail string `json:"userEmail"`
	}

	var order OrderWithUserInfo

	// Get order details
	err = c.db.QueryRow(`
		SELECT o.id, o.user_id, o.total_amount, o.status, o.transaction_id,
			o.shipping_address, o.billing_address, o.payment_method, o.notes,
			o.created_at, o.updated_at,
			u.name as user_name, u.email as user_email
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`, orderID).Scan(
		&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.TransactionID,
		&order.ShippingAddress, &order.BillingAddress, &order.PaymentMethod, &order.Notes,
		&order.CreatedAt, &order.UpdatedAt,
		&order.UserName, &order.UserEmail,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	// Get order items
	rows, err := c.db.Query(`
		SELECT id, product_id, product_name, quantity, unit_price, total_price, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order items"})
		return
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process order items"})
			return
		}

		item.OrderID = orderID
		order.Items = append(order.Items, item)
	}

	// Get payment transactions if available
	var transactions []struct {
		ID           int       `json:"id"`
		TransactionID string   `json:"transactionId"`
		Provider     string    `json:"provider"`
		Amount       float64   `json:"amount"`
		Currency     string    `json:"currency"`
		Status       string    `json:"status"`
		PaymentURL   string    `json:"paymentUrl"`
		Reference    string    `json:"reference"`
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
	}

	tRows, err := c.db.Query(`
		SELECT id, transaction_id, provider, amount, currency, status, payment_url, reference, created_at, updated_at
		FROM payment_transactions
		WHERE order_id = $1
		ORDER BY created_at DESC
	`, orderID)
	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment transactions"})
		return
	}

	if tRows != nil {
		defer tRows.Close()
		for tRows.Next() {
			var t struct {
				ID           int       `json:"id"`
				TransactionID string   `json:"transactionId"`
				Provider     string    `json:"provider"`
				Amount       float64   `json:"amount"`
				Currency     string    `json:"currency"`
				Status       string    `json:"status"`
				PaymentURL   string    `json:"paymentUrl"`
				Reference    string    `json:"reference"`
				CreatedAt    time.Time `json:"createdAt"`
				UpdatedAt    time.Time `json:"updatedAt"`
			}
			err := tRows.Scan(
				&t.ID, &t.TransactionID, &t.Provider, &t.Amount, &t.Currency,
				&t.Status, &t.PaymentURL, &t.Reference, &t.CreatedAt, &t.UpdatedAt,
			)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment transactions"})
				return
			}
			transactions = append(transactions, t)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order": order,
		"payments": transactions,
	})
}

// UpdateOrderStatus updates the status of an order (admin only)
func (c *AdminController) UpdateOrderStatus(ctx *gin.Context) {
	// Get order ID from URL
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Parse request body
	var req struct {
		Status  string `json:"status" binding:"required"`
		TrackingID string `json:"trackingId"`
		Notes   string `json:"notes"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		models.OrderStatusPending:        true,
		models.OrderStatusPaymentPending: true,
		models.OrderStatusPaid:           true,
		models.OrderStatusProcessing:     true,
		models.OrderStatusShipped:        true,
		models.OrderStatusDelivered:      true,
		models.OrderStatusCanceled:       true,
		models.OrderStatusRefunded:       true,
		models.OrderStatusPaymentFailed:  true,
	}

	if !validStatuses[req.Status] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
		return
	}

	// Update order status
	_, err = c.db.Exec(`
		UPDATE orders
		SET status = $1, notes = CASE WHEN $2 <> '' THEN $2 ELSE notes END, updated_at = NOW()
		WHERE id = $3
	`, req.Status, req.Notes, orderID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// If status is "shipped", add tracking ID if provided
	if req.Status == models.OrderStatusShipped && req.TrackingID != "" {
		_, err = c.db.Exec(`
			UPDATE orders
			SET tracking_id = $1
			WHERE id = $2
		`, req.TrackingID, orderID)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracking ID"})
			return
		}
	}

	// Log the status change
	_, err = c.db.Exec(`
		INSERT INTO order_status_history (
			order_id, status, notes, created_at
		) VALUES ($1, $2, $3, NOW())
	`, orderID, req.Status, req.Notes)

	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to log order status change: %v\n", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

// GetOrderStatistics gets order statistics (admin only)
func (c *AdminController) GetOrderStatistics(ctx *gin.Context) {
	// Get period from query params
	period := ctx.DefaultQuery("period", "30days")

	// Define time range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7days":
		startDate = now.AddDate(0, 0, -7)
	case "30days":
		startDate = now.AddDate(0, 0, -30)
	case "90days":
		startDate = now.AddDate(0, 0, -90)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	case "all":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = now.AddDate(0, 0, -30)
	}

	// Get total sales and order count
	var totalSales float64
	var orderCount int
	err := c.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE created_at >= $1
	`, startDate).Scan(&totalSales, &orderCount)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order statistics"})
		return
	}

	// Get sales by status
	type SalesByStatus struct {
		Status string  `json:"status"`
		Count  int     `json:"count"`
		Amount float64 `json:"amount"`
	}

	rows, err := c.db.Query(`
		SELECT status, COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE created_at >= $1
		GROUP BY status
		ORDER BY COUNT(*) DESC
	`, startDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales by status"})
		return
	}
	defer rows.Close()

	var salesByStatus []SalesByStatus
	for rows.Next() {
		var status string
		var count int
		var amount float64
		if err := rows.Scan(&status, &count, &amount); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sales data"})
			return
		}
		salesByStatus = append(salesByStatus, SalesByStatus{
			Status: status,
			Count:  count,
			Amount: amount,
		})
	}

	// Get sales by date
	type SalesByDate struct {
		Date   string  `json:"date"`
		Count  int     `json:"count"`
		Amount float64 `json:"amount"`
	}

	var dateFormat string
	var groupBy string
	if period == "7days" || period == "30days" {
		dateFormat = "YYYY-MM-DD"
		groupBy = "DATE(created_at)"
	} else {
		dateFormat = "YYYY-MM"
		groupBy = "DATE_TRUNC('month', created_at)"
	}

	rows, err = c.db.Query(`
		SELECT TO_CHAR(created_at, $1) as date, COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE created_at >= $2
		GROUP BY date, `+groupBy+`
		ORDER BY `+groupBy+` ASC
	`, dateFormat, startDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales by date"})
		return
	}
	defer rows.Close()

	var salesByDate []SalesByDate
	for rows.Next() {
		var date string
		var count int
		var amount float64
		if err := rows.Scan(&date, &count, &amount); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process date sales data"})
			return
		}
		salesByDate = append(salesByDate, SalesByDate{
			Date:   date,
			Count:  count,
			Amount: amount,
		})
	}

	// Get top selling products
	type TopProduct struct {
		ProductID   int     `json:"productId"`
		ProductName string  `json:"productName"`
		Quantity    int     `json:"quantity"`
		Revenue     float64 `json:"revenue"`
	}

	rows, err = c.db.Query(`
		SELECT product_id, product_name, SUM(quantity) as total_quantity, SUM(total_price) as total_revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE o.created_at >= $1
		GROUP BY product_id, product_name
		ORDER BY total_revenue DESC
		LIMIT 10
	`, startDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top products"})
		return
	}
	defer rows.Close()

	var topProducts []TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.Quantity, &p.Revenue); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process top products data"})
			return
		}
		topProducts = append(topProducts, p)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"totalSales":    totalSales,
		"orderCount":    orderCount,
		"avgOrderValue": func() float64 {
			if orderCount > 0 {
				return totalSales / float64(orderCount)
			}
			return 0
		}(),
		"salesByStatus": salesByStatus,
		"salesByDate":   salesByDate,
		"topProducts":   topProducts,
		"period":        period,
	})
}