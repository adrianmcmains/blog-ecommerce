// api/controllers/admin_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/adrianmcmains/blog-ecommerce/api/models"

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

// GetDashboardData returns aggregated data for the admin dashboard
func (c *AdminController) GetDashboardData(ctx *gin.Context) {
	// Get recent orders
	recentOrdersQuery := `
		SELECT o.id, o.user_id, o.total_amount, o.status, o.created_at,
			   u.name as user_name, u.email as user_email
		FROM orders o
		JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT 5
	`
	
	orderRows, err := c.db.Query(recentOrdersQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent orders"})
		return
	}
	defer orderRows.Close()
	
	var recentOrders []map[string]interface{}
	for orderRows.Next() {
		var order struct {
			ID        int       `json:"id"`
			UserID    int       `json:"userId"`
			Total     float64   `json:"totalAmount"`
			Status    string    `json:"status"`
			CreatedAt time.Time `json:"createdAt"`
			UserName  string    `json:"userName"`
			UserEmail string    `json:"userEmail"`
		}
		err := orderRows.Scan(
			&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt,
			&order.UserName, &order.UserEmail,
		)
		if err != nil {
			continue
		}
		recentOrders = append(recentOrders, map[string]interface{}{
			"id":          order.ID,
			"status":      order.Status,
			"totalAmount": order.Total,
			"createdAt":   order.CreatedAt,
			"customer": map[string]interface{}{
				"id":    order.UserID,
				"name":  order.UserName,
				"email": order.UserEmail,
			},
		})
	}
	
	// Get low stock products
	lowStockQuery := `
		SELECT id, name, sku, stock, price
		FROM products
		WHERE stock <= 5 AND active = true
		ORDER BY stock ASC
		LIMIT 5
	`
	
	productRows, err := c.db.Query(lowStockQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get low stock products"})
		return
	}
	defer productRows.Close()
	
	var lowStockProducts []map[string]interface{}
	for productRows.Next() {
		var product struct {
			ID    int     `json:"id"`
			Name  string  `json:"name"`
			SKU   string  `json:"sku"`
			Stock int     `json:"stock"`
			Price float64 `json:"price"`
		}
		err := productRows.Scan(&product.ID, &product.Name, &product.SKU, &product.Stock, &product.Price)
		if err != nil {
			continue
		}
		lowStockProducts = append(lowStockProducts, map[string]interface{}{
			"id":    product.ID,
			"name":  product.Name,
			"sku":   product.SKU,
			"stock": product.Stock,
			"price": product.Price,
		})
	}
	
	// Get quick stats for today, yesterday, this week, this month
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	
	var todaySales float64
	var todayOrders int
	c.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE DATE(created_at) = $1
	`, today).Scan(&todaySales, &todayOrders)
	
	var yesterdaySales float64
	var yesterdayOrders int
	c.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE DATE(created_at) = $1
	`, yesterday).Scan(&yesterdaySales, &yesterdayOrders)
	
	// Get weekly and monthly stats
	var weeklySales, monthlySales float64
	var weeklyOrders, monthlyOrders int
	
	c.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE)
	`).Scan(&weeklySales, &weeklyOrders)
	
	c.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM orders
		WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&monthlySales, &monthlyOrders)
	
	// Get user stats
	var totalUsers, newUsersToday int
	
	c.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	c.db.QueryRow(`
		SELECT COUNT(*) FROM users
		WHERE DATE(created_at) = $1
	`, today).Scan(&newUsersToday)
	
	ctx.JSON(http.StatusOK, gin.H{
		"recentOrders":      recentOrders,
		"lowStockProducts":  lowStockProducts,
		"quickStats": gin.H{
			"today": gin.H{
				"sales":  todaySales,
				"orders": todayOrders,
			},
			"yesterday": gin.H{
				"sales":  yesterdaySales,
				"orders": yesterdayOrders,
			},
			"thisWeek": gin.H{
				"sales":  weeklySales,
				"orders": weeklyOrders,
			},
			"thisMonth": gin.H{
				"sales":  monthlySales,
				"orders": monthlyOrders,
			},
		},
		"userStats": gin.H{
			"totalUsers":     totalUsers,
			"newUsersToday":  newUsersToday,
		},
	})
}

// GetAllUsers returns a list of all users
func (c *AdminController) GetAllUsers(ctx *gin.Context) {
	role := ctx.Query("role")
	search := ctx.Query("search")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	
	// Build the query
	query := `
		SELECT id, email, name, role, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	
	// Add filters
	var args []interface{}
	argCount := 1
	
	if role != "" {
		query += " AND role = $" + strconv.Itoa(argCount)
		args = append(args, role)
		argCount++
	}
	
	if search != "" {
		query += " AND (email ILIKE $" + strconv.Itoa(argCount) + " OR name ILIKE $" + strconv.Itoa(argCount+1) + ")"
		args = append(args, "%"+search+"%")
		argCount += 2
	}
	
	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)
	
	// Execute the query
	rows, err := c.db.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()
	
	// Prepare result
	var users []map[string]interface{}
	for rows.Next() {
		var user struct {
			ID        int       `json:"id"`
			Email     string    `json:"email"`
			Name      string    `json:"name"`
			Role      string    `json:"role"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
		}
		
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		
		users = append(users, map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		})
	}
	
	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	
	// Add WHERE clause if needed
	if role != "" {
		countQuery += " AND role = $1"
	}
	
	if search != "" {
		if role != "" {
			countQuery += " AND (email ILIKE $2 OR name ILIKE $2)"
		} else {
			countQuery += " AND (email ILIKE $1 OR name ILIKE $1)"
		}
	}
	
	var countArgs []interface{}
	if role != "" {
		countArgs = append(countArgs, role)
	}
	if search != "" {
		countArgs = append(countArgs, "%"+search+"%")
	}
	
	var total int
	c.db.QueryRow(countQuery, countArgs...).Scan(&total)
	
	ctx.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser gets a single user by ID
func (c *AdminController) GetUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user struct {
		ID        int       `json:"id"`
		Email     string    `json:"email"`
		Name      string    `json:"name"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	
	err = c.db.QueryRow(`
		SELECT id, email, name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	// Get user's order count and total spend
	var orderCount int
	var totalSpend float64
	
	c.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE user_id = $1
	`, userID).Scan(&orderCount, &totalSpend)
	
	// Get user's addresses
	rows, err := c.db.Query(`
		SELECT id, address_line1, address_line2, city, state, postal_code, country, is_default
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`, userID)
	
	var addresses []map[string]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var addr struct {
				ID          int     `json:"id"`
				Line1       string  `json:"addressLine1"`
				Line2       *string `json:"addressLine2"`
				City        string  `json:"city"`
				State       string  `json:"state"`
				PostalCode  string  `json:"postalCode"`
				Country     string  `json:"country"`
				IsDefault   bool    `json:"isDefault"`
			}
			
			err := rows.Scan(&addr.ID, &addr.Line1, &addr.Line2, &addr.City, &addr.State, 
				&addr.PostalCode, &addr.Country, &addr.IsDefault)
			if err != nil {
				continue
			}
			
			addresses = append(addresses, map[string]interface{}{
				"id":           addr.ID,
				"addressLine1": addr.Line1,
				"addressLine2": addr.Line2,
				"city":         addr.City,
				"state":        addr.State,
				"postalCode":   addr.PostalCode,
				"country":      addr.Country,
				"isDefault":    addr.IsDefault,
			})
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
			"stats": map[string]interface{}{
				"orderCount": orderCount,
				"totalSpend": totalSpend,
			},
			"addresses": addresses,
		},
	})
}

// UpdateUser updates a user's information
func (c *AdminController) UpdateUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if email is already in use by another user
	if req.Email != "" {
		var count int
		err := c.db.QueryRow(`
			SELECT COUNT(*) FROM users
			WHERE email = $1 AND id != $2
		`, req.Email, userID).Scan(&count)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email uniqueness"})
			return
		}
		
		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email is already in use"})
			return
		}
	}
	
	// Update user
	_, err = c.db.Exec(`
		UPDATE users
		SET name = CASE WHEN $1 <> '' THEN $1 ELSE name END,
			email = CASE WHEN $2 <> '' THEN $2 ELSE email END,
			updated_at = NOW()
		WHERE id = $3
	`, req.Name, req.Email, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user
func (c *AdminController) DeleteUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Check if user is an admin
	var role string
	err = c.db.QueryRow(`SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if role == "admin" {
		// Count how many admins we have
		var adminCount int
		c.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount)
		
		if adminCount <= 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the only admin"})
			return
		}
	}
	
	// Start transaction for proper deletion
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	// Delete user addresses
	_, err = tx.Exec(`DELETE FROM user_addresses WHERE user_id = $1`, userID)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user addresses"})
		return
	}
	
	// Delete user cart items
	_, err = tx.Exec(`
		DELETE FROM cart_items 
		WHERE cart_id IN (SELECT id FROM carts WHERE user_id = $1)
	`, userID)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart items"})
		return
	}
	
	// Delete user cart
	_, err = tx.Exec(`DELETE FROM carts WHERE user_id = $1`, userID)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
		return
	}
	
	// Delete user wishlist
	_, err = tx.Exec(`DELETE FROM wishlist_items WHERE user_id = $1`, userID)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete wishlist"})
		return
	}
	
	// Check if user has orders
	var orderCount int
	err = tx.QueryRow(`SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&orderCount)
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check orders"})
		return
	}
	
	if orderCount > 0 {
		// If user has orders, we might want to anonymize instead of delete
		_, err = tx.Exec(`
			UPDATE users
			SET email = CONCAT('deleted_', id, '@example.com'),
				name = 'Deleted User',
				password_hash = NULL,
				active = false,
				deleted_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
		`, userID)
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to anonymize user"})
			return
		}
	} else {
		// If no orders, we can delete the user
		_, err = tx.Exec(`DELETE FROM users WHERE id = $1`, userID)
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ChangeUserRole changes a user's role
func (c *AdminController) ChangeUserRole(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate role
	validRoles := map[string]bool{
		"admin":       true,
		"customer":    true,
		"contributor": true,
	}
	
	if !validRoles[req.Role] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	
	// If changing from admin, check if this is the last admin
	if req.Role != "admin" {
		var currentRole string
		err := c.db.QueryRow(`SELECT role FROM users WHERE id = $1`, userID).Scan(&currentRole)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user role"})
			}
			return
		}
		
		if currentRole == "admin" {
			var adminCount int
			c.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount)
			
			if adminCount <= 1 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the only admin role"})
				return
			}
		}
	}
	
	// Update user role
	result, err := c.db.Exec(`
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2
	`, req.Role, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// GetSystemLogs gets system logs
func (c *AdminController) GetSystemLogs(ctx *gin.Context) {
	logType := ctx.DefaultQuery("type", "all")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	
	query := `
		SELECT id, log_type, message, metadata, created_at
		FROM system_logs
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1
	
	if logType != "all" {
		query += " AND log_type = $" + strconv.Itoa(argCount)
		args = append(args, logType)
		argCount++
	}
	
	query += " ORDER BY created_at DESC"
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)
	
	rows, err := c.db.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	defer rows.Close()
	
	var logs []map[string]interface{}
	for rows.Next() {
		var log struct {
			ID        int       `json:"id"`
			LogType   string    `json:"logType"`
			Message   string    `json:"message"`
			Metadata  string    `json:"metadata"`
			CreatedAt time.Time `json:"createdAt"`
		}
		
		err := rows.Scan(&log.ID, &log.LogType, &log.Message, &log.Metadata, &log.CreatedAt)
		if err != nil {
			continue
		}
		
		// Parse metadata as JSON if possible
		var metadataJSON interface{}
		if err := json.Unmarshal([]byte(log.Metadata), &metadataJSON); err != nil {
			metadataJSON = log.Metadata
		}
		
		logs = append(logs, map[string]interface{}{
			"id":        log.ID,
			"logType":   log.LogType,
			"message":   log.Message,
			"metadata":  metadataJSON,
			"createdAt": log.CreatedAt,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// GetSiteStatistics gets overall site statistics
func (c *AdminController) GetSiteStatistics(ctx *gin.Context) {
	// Get user stats
	var totalUsers, activeUsers int
	c.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totalUsers)
	c.db.QueryRow(`
		SELECT COUNT(*) FROM users
		WHERE last_login_at >= NOW() - INTERVAL '30 days'
	`).Scan(&activeUsers)
	
	// Get content stats
	var totalProducts, totalPosts, totalCategories int
	c.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&totalProducts)
	c.db.QueryRow(`SELECT COUNT(*) FROM blog_posts`).Scan(&totalPosts)
	c.db.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT id FROM product_categories
			UNION
			SELECT id FROM blog_categories
		) as categories
	`).Scan(&totalCategories)
	
	// Get order stats
	var totalOrders, pendingOrders int
	var totalRevenue float64
	c.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(total_amount), 0) FROM orders`).Scan(&totalOrders, &totalRevenue)
	c.db.QueryRow(`
		SELECT COUNT(*) FROM orders
		WHERE status IN ('pending', 'processing')
	`).Scan(&pendingOrders)
	
	// Get traffic stats if available
	var pageViews, uniqueVisitors int
	c.db.QueryRow(`
		SELECT 
			COALESCE(SUM(page_views), 0),
			COUNT(DISTINCT visitor_id)
		FROM site_analytics
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`).Scan(&pageViews, &uniqueVisitors)
	
	ctx.JSON(http.StatusOK, gin.H{
		"userStats": map[string]interface{}{
			"totalUsers":   totalUsers,
			"activeUsers":  activeUsers,
			"userGrowth":   0, // Would require historical data
		},
		"contentStats": map[string]interface{}{
			"totalProducts":   totalProducts,
			"totalPosts":      totalPosts,
			"totalCategories": totalCategories,
		},
		"orderStats": map[string]interface{}{
			"totalOrders":    totalOrders,
			"pendingOrders":  pendingOrders,
			"totalRevenue":   totalRevenue,
		},
		"trafficStats": map[string]interface{}{
			"pageViews":      pageViews,
			"uniqueVisitors": uniqueVisitors,
		},
	})
}

// GetSystemSettings gets system settings
func (c *AdminController) GetSystemSettings(ctx *gin.Context) {
	rows, err := c.db.Query(`
		SELECT setting_key, setting_value, setting_type
		FROM system_settings
		ORDER BY setting_key
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	defer rows.Close()
	
	settings := make(map[string]interface{})
	for rows.Next() {
		var key, value, settingType string
		err := rows.Scan(&key, &value, &settingType)
		if err != nil {
			continue
		}
		
		switch settingType {
		case "boolean":
			settings[key] = value == "true"
		case "number":
			if numVal, err := strconv.ParseFloat(value, 64); err == nil {
				settings[key] = numVal
			} else {
				settings[key] = value
			}
		case "json":
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(value), &jsonVal); err == nil {
				settings[key] = jsonVal
			} else {
				settings[key] = value
			}
		default:
			settings[key] = value
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{"settings": settings})
}

// UpdateSystemSettings updates system settings
func (c *AdminController) UpdateSystemSettings(ctx *gin.Context) {
	var req map[string]interface{}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	for key, value := range req {
		// Determine setting type
		var settingType string
		var stringValue string
		
		switch v := value.(type) {
		case bool:
			settingType = "boolean"
			if v {
				stringValue = "true"
			} else {
				stringValue = "false"
			}
		case float64:
			settingType = "number"
			stringValue = strconv.FormatFloat(v, 'f', -1, 64)
		case map[string]interface{}, []interface{}:
			settingType = "json"
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				continue
			}
			stringValue = string(jsonBytes)
		default:
			settingType = "string"
			if str, ok := value.(string); ok {
				stringValue = str
			} else {
				continue
			}
		}
		
		// Upsert setting
		_, err := tx.Exec(`
			INSERT INTO system_settings (setting_key, setting_value, setting_type, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (setting_key) 
			DO UPDATE SET setting_value = $2, setting_type = $3, updated_at = NOW()
		`, key, stringValue, settingType)
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// BackupDatabase creates a database backup
func (c *AdminController) BackupDatabase(ctx *gin.Context) {
	// Get database connection details from environment or config
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	
	if dbName == "" || dbUser == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database configuration is incomplete"})
		return
	}
	
	// Create backup directory if it doesn't exist
	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup directory"})
		return
	}
	
	// Create a timestamp for the backup file
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))
	
	// Build pg_dump command
	pgDumpCmd := exec.Command("pg_dump", 
		"-h", dbHost,
		"-p", dbPort,
		"-U", dbUser,
		"-d", dbName,
		"-f", backupFile,
		"-F", "c")  // Custom format for compressed output
	
	// Set password environment variable
	pgDumpCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbPassword))
	
	// Execute the command
	output, err := pgDumpCmd.CombinedOutput()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create database backup",
			"details": string(output),
		})
		return
	}
	
	// Log the backup
	_, err = c.db.Exec(`
		INSERT INTO system_logs (log_type, message, metadata, created_at)
		VALUES ($1, $2, $3, NOW())
	`, "backup", "Database backup created", fmt.Sprintf(`{"file":"%s"}`, backupFile))
	
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to log database backup: %v\n", err)
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Database backup created successfully",
		"file": backupFile,
	})
}

// RestoreDatabase restores a database from backup
func (c *AdminController) RestoreDatabase(ctx *gin.Context) {
	var req struct {
		BackupFile string `json:"backupFile" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate backup file path
	backupFile := req.BackupFile
	if !filepath.IsAbs(backupFile) {
		backupFile = filepath.Join("./backups", backupFile)
	}
	
	// Check if file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Backup file does not exist"})
		return
	}
	
	// Get database connection details
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	
	if dbName == "" || dbUser == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database configuration is incomplete"})
		return
	}
	
	// Build pg_restore command
	pgRestoreCmd := exec.Command("pg_restore", 
		"-h", dbHost,
		"-p", dbPort,
		"-U", dbUser,
		"-d", dbName,
		"-c",  // Clean (drop) database objects before recreating
		"-v",  // Verbose mode
		backupFile)
	
	// Set password environment variable
	pgRestoreCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbPassword))
	
	// Execute the command
	output, err := pgRestoreCmd.CombinedOutput()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restore database backup",
			"details": string(output),
		})
		return
	}
	
	// Log the restore
	_, err = c.db.Exec(`
		INSERT INTO system_logs (log_type, message, metadata, created_at)
		VALUES ($1, $2, $3, NOW())
	`, "restore", "Database restored from backup", fmt.Sprintf(`{"file":"%s"}`, backupFile))
	
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to log database restore: %v\n", err)
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Database restored successfully",
	})
}

// GetSystemHealth checks system health
func (c *AdminController) GetSystemHealth(ctx *gin.Context) {
	// Check database connection
	dbStatus := "ok"
	dbErr := c.db.Ping()
	if dbErr != nil {
		dbStatus = "error"
	}
	
	// Check disk space
	diskStatus := "ok"
	var diskFree uint64
	var diskTotal uint64
	var diskUsedPercent float64
	
	// This is simplified and would need proper implementation in production
	diskFree = 1000000000  // 1 GB free space
	diskTotal = 10000000000  // 10 GB total space
	diskUsedPercent = 90.0  // 90% used
	
	if diskUsedPercent > 90 {
		diskStatus = "warning"
	}
	
	// Check memory usage
	memStatus := "ok"
	var memFree uint64
	var memTotal uint64
	var memUsedPercent float64
	
	// This is simplified and would need proper implementation in production
	memFree = 100000000  // 100 MB free
	memTotal = 1000000000  // 1 GB total
	memUsedPercent = 90.0  // 90% used
	
	if memUsedPercent > 90 {
		memStatus = "warning"
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"status": map[string]string{
			"overall": func() string {
				if dbStatus == "error" {
					return "error"
				}
				if diskStatus == "warning" || memStatus == "warning" {
					return "warning"
				}
				return "ok"
			}(),
			"database": dbStatus,
			"disk":     diskStatus,
			"memory":   memStatus,
		},
		"details": map[string]interface{}{
			"database": map[string]interface{}{
				"connected": dbErr == nil,
				"error":     func() string {
					if dbErr != nil {
						return dbErr.Error()
					}
					return ""
				}(),
			},
			"disk": map[string]interface{}{
				"total":       diskTotal,
				"free":        diskFree,
				"usedPercent": diskUsedPercent,
			},
			"memory": map[string]interface{}{
				"total":       memTotal,
				"free":        memFree,
				"usedPercent": memUsedPercent,
			},
			"uptime": map[string]interface{}{
				"server": "Unknown", // Would need to be implemented
				"app":    "Unknown", // Would need to be implemented
			},
		},
	})
}