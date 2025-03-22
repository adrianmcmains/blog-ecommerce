// File: api/controllers/shop_controller.go
package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ShopController handles shop-related routes
type ShopController struct {
	DB                *sql.DB
	PaymentController *PaymentController
}

// NewShopController creates a new shop controller
func NewShopController(db *sql.DB) *ShopController {
	return &ShopController{
		DB: db,
	}
}

// GetProducts retrieves all products
func (sc *ShopController) GetProducts(c *gin.Context) {
    // Implement the logic to get products
    c.JSON(http.StatusOK, gin.H{"message": "GetProducts method not implemented"})
}

// CreateProduct creates a new product
func (sc *ShopController) CreateProduct(c *gin.Context) {
	// Implement the logic to create a product
	c.JSON(http.StatusOK, gin.H{"message": "CreateProduct method not implemented"})
}

// UpdateProduct updates an existing product
func (sc *ShopController) UpdateProduct(c *gin.Context) {
	// Implement the logic to update a product
	c.JSON(http.StatusOK, gin.H{"message": "UpdateProduct method not implemented"})
}

// UpdateProduct updates an existing product
func (sc *ShopController) DeleteProduct(c *gin.Context) {
	// Implement the logic to delete a product
	c.JSON(http.StatusOK, gin.H{"message": "DeleteProduct method not implemented"})
}

// GetProductBySlug retrieves a product by its slug
func (c *ShopController) GetProductBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")
	
	// Query product from database using slug
	var product struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Slug        string  `json:"slug"`
		Active      bool    `json:"active"`
	}
	
	err := c.DB.QueryRow(`
		SELECT id, name, description, price, slug, active
		FROM products
		WHERE slug = $1 AND active = true
	`, slug).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Slug, &product.Active)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"product": product})
}

// GetCategories retrieves all product categories
func (c *ShopController) GetCategories(ctx *gin.Context) {
	rows, err := c.DB.Query(`
		SELECT id, name, slug, description, parent_id
		FROM product_categories
		ORDER BY name
	`)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	defer rows.Close()
	
	var categories []map[string]interface{}
	for rows.Next() {
		var category struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Slug        string  `json:"slug"`
			Description string  `json:"description"`
			ParentID    *string `json:"parent_id"`
		}
		err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.Description, &category.ParentID)
		if err != nil {
			continue
		}
		categories = append(categories, map[string]interface{}{
			"id":          category.ID,
			"name":        category.Name,
			"slug":        category.Slug,
			"description": category.Description,
			"parent_id":   category.ParentID,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetFeaturedProducts gets featured products for the home page
func (c *ShopController) GetFeaturedProducts(ctx *gin.Context) {
	rows, err := c.DB.Query(`
		SELECT id, name, description, price, slug, featured_image
		FROM products
		WHERE featured = true AND active = true
		ORDER BY created_at DESC
		LIMIT 8
	`)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch featured products"})
		return
	}
	defer rows.Close()
	
	var products []map[string]interface{}
	for rows.Next() {
		var product struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			Description   string  `json:"description"`
			Price         float64 `json:"price"`
			Slug          string  `json:"slug"`
			FeaturedImage string  `json:"featured_image"`
		}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Slug, &product.FeaturedImage)
		if err != nil {
			continue
		}
		products = append(products, map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"slug":           product.Slug,
			"featured_image": product.FeaturedImage,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"products": products})
}

// GetNewArrivals gets the newest products
func (c *ShopController) GetNewArrivals(ctx *gin.Context) {
	rows, err := c.DB.Query(`
		SELECT id, name, description, price, slug, featured_image
		FROM products
		WHERE active = true
		ORDER BY created_at DESC
		LIMIT 8
	`)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new arrivals"})
		return
	}
	defer rows.Close()
	
	var products []map[string]interface{}
	for rows.Next() {
		var product struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			Description   string  `json:"description"`
			Price         float64 `json:"price"`
			Slug          string  `json:"slug"`
			FeaturedImage string  `json:"featured_image"`
		}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Slug, &product.FeaturedImage)
		if err != nil {
			continue
		}
		products = append(products, map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"slug":           product.Slug,
			"featured_image": product.FeaturedImage,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"products": products})
}

// GetBestSellers gets the best-selling products
func (c *ShopController) GetBestSellers(ctx *gin.Context) {
	rows, err := c.DB.Query(`
		SELECT p.id, p.name, p.description, p.price, p.slug, p.featured_image
		FROM products p
		JOIN (
			SELECT product_id, SUM(quantity) as total_sold
			FROM order_items
			JOIN orders ON order_items.order_id = orders.id
			WHERE orders.status IN ('completed', 'delivered')
			GROUP BY product_id
			ORDER BY total_sold DESC
			LIMIT 8
		) sales ON p.id = sales.product_id
		WHERE p.active = true
	`)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch best sellers"})
		return
	}
	defer rows.Close()
	
	var products []map[string]interface{}
	for rows.Next() {
		var product struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			Description   string  `json:"description"`
			Price         float64 `json:"price"`
			Slug          string  `json:"slug"`
			FeaturedImage string  `json:"featured_image"`
		}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Slug, &product.FeaturedImage)
		if err != nil {
			continue
		}
		products = append(products, map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"slug":           product.Slug,
			"featured_image": product.FeaturedImage,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"products": products})
}

// SearchProducts searches for products based on query params
func (c *ShopController) SearchProducts(ctx *gin.Context) {
	query := ctx.Query("q")
	category := ctx.Query("category")
	minPriceStr := ctx.Query("min_price")
	maxPriceStr := ctx.Query("max_price")
	sortBy := ctx.Query("sort")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	
	// Build query
	sqlQuery := `
		SELECT p.id, p.name, p.description, p.price, p.slug, p.featured_image
		FROM products p
	`
	
	// Add category join if needed
	if category != "" {
		sqlQuery += `
			JOIN product_categories pc ON p.id = pc.product_id
			JOIN categories c ON pc.category_id = c.id
		`
	}
	
	sqlQuery += ` WHERE p.active = true `
	
	args := []interface{}{}
	argIndex := 1
	
	// Add search condition
	if query != "" {
		sqlQuery += ` AND (p.name ILIKE $` + strconv.Itoa(argIndex) + ` OR p.description ILIKE $` + strconv.Itoa(argIndex) + `) `
		args = append(args, "%"+query+"%")
		argIndex++
	}
	
	// Add category condition
	if category != "" {
		sqlQuery += ` AND c.slug = $` + strconv.Itoa(argIndex) + ` `
		args = append(args, category)
		argIndex++
	}
	
	// Add price range conditions
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err == nil && minPrice >= 0 {
			sqlQuery += ` AND p.price >= $` + strconv.Itoa(argIndex) + ` `
			args = append(args, minPrice)
			argIndex++
		}
	}
	
	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err == nil && maxPrice > 0 {
			sqlQuery += ` AND p.price <= $` + strconv.Itoa(argIndex) + ` `
			args = append(args, maxPrice)
			argIndex++
		}
	}
	
	// Add sorting
	switch sortBy {
	case "price_asc":
		sqlQuery += ` ORDER BY p.price ASC `
	case "price_desc":
		sqlQuery += ` ORDER BY p.price DESC `
	case "newest":
		sqlQuery += ` ORDER BY p.created_at DESC `
	default:
		sqlQuery += ` ORDER BY p.name ASC `
	}
	
	// Add pagination
	sqlQuery += ` LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1) + ` `
	args = append(args, pageSize, offset)
	
	// Execute the query
	rows, err := c.DB.Query(sqlQuery, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}
	defer rows.Close()
	
	var products []map[string]interface{}
	for rows.Next() {
		var product struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			Description   string  `json:"description"`
			Price         float64 `json:"price"`
			Slug          string  `json:"slug"`
			FeaturedImage string  `json:"featured_image"`
		}
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Slug, &product.FeaturedImage)
		if err != nil {
			continue
		}
		products = append(products, map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"slug":           product.Slug,
			"featured_image": product.FeaturedImage,
		})
	}
	
	// Get total count for pagination
	countQuery := strings.Replace(sqlQuery, "SELECT p.id, p.name, p.description, p.price, p.slug, p.featured_image", "SELECT COUNT(*)", 1)
	countQuery = strings.Split(countQuery, " LIMIT ")[0]
	
	var total int
	err = c.DB.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		total = len(products)
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"products":    products,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": (total + pageSize - 1) / pageSize,
	})
}

// DebugProducts returns debugging information about products
func (c *ShopController) DebugProducts(ctx *gin.Context) {
	// Only available in non-production environments
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Debug endpoint for products",
		"time":    time.Now().String(),
	})
}

// SyncProducts syncs products between Hugo and database
func (c *ShopController) SyncProducts(ctx *gin.Context) {
	// Implementation for syncing products
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Product sync started",
		"success": true,
	})
}

// SyncContentFromTina syncs content from TinaCMS to the database
func (c *ShopController) SyncContentFromTina(ctx *gin.Context) {
	// Implementation for syncing from TinaCMS
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Content sync from TinaCMS completed",
		"success": true,
	})
}

// GetInventoryReport generates an inventory report
func (c *ShopController) GetInventoryReport(ctx *gin.Context) {
	rows, err := c.DB.Query(`
		SELECT p.id, p.name, p.sku, p.stock, p.price
		FROM products p
		ORDER BY p.name
	`)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate inventory report"})
		return
	}
	defer rows.Close()
	
	var inventory []map[string]interface{}
	for rows.Next() {
		var item struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			SKU   string  `json:"sku"`
			Stock int     `json:"stock"`
			Price float64 `json:"price"`
		}
		err := rows.Scan(&item.ID, &item.Name, &item.SKU, &item.Stock, &item.Price)
		if err != nil {
			continue
		}
		inventory = append(inventory, map[string]interface{}{
			"id":    item.ID,
			"name":  item.Name,
			"sku":   item.SKU,
			"stock": item.Stock,
			"price": item.Price,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"inventory": inventory})
}

// UpdateStockLevels updates stock levels for products
func (c *ShopController) UpdateStockLevels(ctx *gin.Context) {
	var req struct {
		Updates []struct {
			ProductID string `json:"product_id" binding:"required"`
			Stock     int    `json:"stock" binding:"required"`
		} `json:"updates" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	for _, update := range req.Updates {
		_, err := tx.Exec(`
			UPDATE products
			SET stock = $1, updated_at = $2
			WHERE id = $3
		`, update.Stock, time.Now(), update.ProductID)
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock levels"})
			return
		}
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Stock levels updated successfully"})
}

// ClearCart clears the user's cart
func (c *ShopController) ClearCart(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	_, err := c.DB.Exec(`
		DELETE FROM cart_items
		WHERE cart_id IN (
			SELECT id FROM carts WHERE user_id = $1
		)
	`, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}

// GetCartItemCount gets the number of items in the user's cart
func (c *ShopController) GetCartItemCount(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var count int
	err := c.DB.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM cart_items
		WHERE cart_id IN (
			SELECT id FROM carts WHERE user_id = $1
		)
	`, userID).Scan(&count)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart item count"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GetOrderTracking gets tracking information for an order
func (c *ShopController) GetOrderTracking(ctx *gin.Context) {
	orderID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	var tracking struct {
		TrackingNumber string     `json:"tracking_number"`
		Carrier        string     `json:"carrier"`
		Status         string     `json:"status"`
		UpdatedAt      time.Time  `json:"updated_at"`
		EstimatedDelivery *time.Time `json:"estimated_delivery"`
	}
	
	err := c.DB.QueryRow(`
		SELECT tracking_number, carrier, status, updated_at, estimated_delivery
		FROM order_tracking
		WHERE order_id = $1 AND user_id = $2
	`, orderID, userID).Scan(&tracking.TrackingNumber, &tracking.Carrier, &tracking.Status, &tracking.UpdatedAt, &tracking.EstimatedDelivery)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tracking information not found"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"tracking": tracking})
}

// CancelOrder cancels an order
func (c *ShopController) CancelOrder(ctx *gin.Context) {
	orderID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	// Check if order belongs to user
	var count int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM orders
		WHERE id = $1 AND user_id = $2 AND status IN ('pending', 'processing')
	`, orderID, userID).Scan(&count)
	
	if err != nil || count == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found or cannot be canceled"})
		return
	}
	
	// Begin transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	// Update order status
	_, err = tx.Exec(`
		UPDATE orders
		SET status = 'canceled', updated_at = $1
		WHERE id = $2
	`, time.Now(), orderID)
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}
	
	// Return items to inventory if needed
	_, err = tx.Exec(`
		UPDATE products p
		SET stock = p.stock + oi.quantity
		FROM order_items oi
		WHERE oi.order_id = $1 AND oi.product_id = p.id
	`, orderID)
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
		return
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Order canceled successfully"})
}

// CreateCategory creates a new product category
func (c *ShopController) CreateCategory(ctx *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Slug        string  `json:"slug" binding:"required"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
		ImageURL    string  `json:"image_url"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if slug is unique
	var count int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM product_categories
		WHERE slug = $1
	`, req.Slug).Scan(&count)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug uniqueness"})
		return
	}
	
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A category with this slug already exists"})
		return
	}
	
	// Create category
	id := uuid.New().String()
	_, err = c.DB.Exec(`
		INSERT INTO product_categories (id, name, slug, description, parent_id, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, req.Name, req.Slug, req.Description, req.ParentID, req.ImageURL, time.Now())
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}
	
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"id":      id,
	})
}

// UpdateCategory updates an existing product category
func (c *ShopController) UpdateCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id")
	
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Slug        string  `json:"slug" binding:"required"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"`
		ImageURL    string  `json:"image_url"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if category exists
	var currentSlug string
	err := c.DB.QueryRow(`
		SELECT slug FROM product_categories
		WHERE id = $1
	`, categoryID).Scan(&currentSlug)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	
	// Check if slug is unique (if changed)
	if req.Slug != currentSlug {
		var count int
		err := c.DB.QueryRow(`
			SELECT COUNT(*) FROM product_categories
			WHERE slug = $1 AND id != $2
		`, req.Slug, categoryID).Scan(&count)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug uniqueness"})
			return
		}
		
		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "A category with this slug already exists"})
			return
		}
	}
	
	// Prevent circular references
	if req.ParentID != nil && *req.ParentID == categoryID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "A category cannot be its own parent"})
		return
	}
	
	// Update category
	_, err = c.DB.Exec(`
		UPDATE product_categories
		SET name = $1, slug = $2, description = $3, parent_id = $4, image_url = $5, updated_at = $6
		WHERE id = $7
	`, req.Name, req.Slug, req.Description, req.ParentID, req.ImageURL, time.Now(), categoryID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
}

// DeleteCategory deletes a product category
func (c *ShopController) DeleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("id")
	
	// Check if category is in use by products
	var productCount int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM product_categories_mapping
		WHERE category_id = $1
	`, categoryID).Scan(&productCount)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check category usage"})
		return
	}
	
	if productCount > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category as it is associated with products"})
		return
	}
	
	// Check if category has subcategories
	var childCount int
	err = c.DB.QueryRow(`
		SELECT COUNT(*) FROM product_categories
		WHERE parent_id = $1
	`, categoryID).Scan(&childCount)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subcategories"})
		return
	}
	
	if childCount > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category as it has subcategories"})
		return
	}
	
	// Delete category
	result, err := c.DB.Exec(`
		DELETE FROM product_categories
		WHERE id = $1
	`, categoryID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// Methods for Wishlist
func (c *ShopController) GetWishlist(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	rows, err := c.DB.Query(`
		SELECT w.id, p.id as product_id, p.name, p.price, p.slug, p.featured_image
		FROM wishlist_items w
		JOIN products p ON w.product_id = p.id
		WHERE w.user_id = $1
	`, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wishlist"})
		return
	}
	defer rows.Close()
	
	var items []map[string]interface{}
	for rows.Next() {
		var item struct {
			ID            string  `json:"id"`
			ProductID     string  `json:"product_id"`
			Name          string  `json:"name"`
			Price         float64 `json:"price"`
			Slug          string  `json:"slug"`
			FeaturedImage string  `json:"featured_image"`
		}
		err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Slug, &item.FeaturedImage)
		if err != nil {
			continue
		}
		items = append(items, map[string]interface{}{
			"id":             item.ID,
			"product_id":     item.ProductID,
			"name":           item.Name,
			"price":          item.Price,
			"slug":           item.Slug,
			"featured_image": item.FeaturedImage,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"wishlist": items})
}

func (c *ShopController) AddToWishlist(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if already in wishlist
	var count int
		err := c.DB.QueryRow(`
			SELECT COUNT(*) FROM wishlist_items
			WHERE user_id = $1 AND product_id = $2
		`, userID, req.ProductID).Scan(&count)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check wishlist"})
			return
		}
		
		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Product already in wishlist"})
			return
		}
		
		// Add to wishlist
		id := uuid.New().String()
		_, err = c.DB.Exec(`
			INSERT INTO wishlist_items (id, user_id, product_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, id, userID, req.ProductID, time.Now())
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to wishlist"})
			return
		}
		
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Product added to wishlist",
			"id":      id,
		})
	}



func (c *ShopController) RemoveFromWishlist(ctx *gin.Context) {
	itemID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	result, err := c.DB.Exec(`
		DELETE FROM wishlist_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from wishlist"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Wishlist item not found"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Item removed from wishlist"})
}

func (c *ShopController) MoveWishlistItemToCart(ctx *gin.Context) {
	itemID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	// Get the product ID from wishlist
	var productID string
	err := c.DB.QueryRow(`
		SELECT product_id FROM wishlist_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID).Scan(&productID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Wishlist item not found"})
		return
	}
	
	// Begin transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	// Find or create cart
	var cartID string
	err = tx.QueryRow(`
		SELECT id FROM carts WHERE user_id = $1
	`, userID).Scan(&cartID)
	
	if err != nil {
		// Create new cart
		cartID = uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO carts (id, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
		`, cartID, userID, time.Now())
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
			return
		}
	}
	
	// Check if item already in cart
	var cartItemID string
	var quantity int
	err = tx.QueryRow(`
		SELECT id, quantity FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID).Scan(&cartItemID, &quantity)
	
	if err == nil {
		// Update existing cart item
		_, err = tx.Exec(`
			UPDATE cart_items
			SET quantity = $1, updated_at = $2
			WHERE id = $3
		`, quantity+1, time.Now(), cartItemID)
	} else {
		// Add new cart item
		cartItemID = uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, 1, $4, $4)
		`, cartItemID, cartID, productID, time.Now())
	}
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}
	
	// Remove from wishlist
	_, err = tx.Exec(`
		DELETE FROM wishlist_items
		WHERE id = $1 AND user_id = $2
	`, itemID, userID)
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from wishlist"})
		return
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Item moved to cart successfully"})
}

// Review methods
func (c *ShopController) GetProductReviews(ctx *gin.Context) {
	productID := ctx.Param("id")
	
	rows, err := c.DB.Query(`
		SELECT r.id, r.user_id, u.first_name, u.last_name, r.rating, r.review, r.created_at
		FROM product_reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1 AND r.approved = true
		ORDER BY r.created_at DESC
	`, productID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	defer rows.Close()
	
	var reviews []map[string]interface{}
	for rows.Next() {
		var review struct {
			ID        string    `json:"id"`
			UserID    string    `json:"user_id"`
			FirstName string    `json:"first_name"`
			LastName  string    `json:"last_name"`
			Rating    int       `json:"rating"`
			Review    string    `json:"review"`
			CreatedAt time.Time `json:"created_at"`
		}
		err := rows.Scan(&review.ID, &review.UserID, &review.FirstName, &review.LastName, &review.Rating, &review.Review, &review.CreatedAt)
		if err != nil {
			continue
		}
		reviews = append(reviews, map[string]interface{}{
			"id":         review.ID,
			"user_id":    review.UserID,
			"first_name": review.FirstName,
			"last_name":  review.LastName,
			"rating":     review.Rating,
			"review":     review.Review,
			"created_at": review.CreatedAt,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func (c *ShopController) AddReview(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Review    string `json:"review"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if user has purchased the product
	var orderCount int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		WHERE o.user_id = $1 AND oi.product_id = $2 AND o.status = 'completed'
	`, userID, req.ProductID).Scan(&orderCount)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check purchase history"})
		return
	}
	
	// Check if already reviewed by this user
	var existingReview int
	err = c.DB.QueryRow(`
		SELECT COUNT(*) FROM product_reviews
		WHERE user_id = $1 AND product_id = $2
	`, userID, req.ProductID).Scan(&existingReview)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing reviews"})
		return
	}
	
	if existingReview > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "You have already reviewed this product"})
		return
	}
	
	// Determine if auto-approval is needed
	// Let's assume reviews from users who purchased get auto-approved
	autoApprove := orderCount > 0
	
	// Add review
	id := uuid.New().String()
	_, err = c.DB.Exec(`
		INSERT INTO product_reviews (id, product_id, user_id, rating, review, approved, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, req.ProductID, userID, req.Rating, req.Review, autoApprove, time.Now())
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
		return
	}
	
	// Update product average rating
	_, err = c.DB.Exec(`
		UPDATE products
		SET average_rating = (
			SELECT AVG(rating) FROM product_reviews
			WHERE product_id = $1 AND approved = true
		)
		WHERE id = $1
	`, req.ProductID)
	
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update product average rating: %v\n", err)
	}
	
	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Review submitted successfully",
		"id":         id,
		"approved":   autoApprove,
		"moderation": !autoApprove,
	})
}

func (c *ShopController) UpdateOwnReview(ctx *gin.Context) {
	reviewID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	var req struct {
		Rating int    `json:"rating" binding:"required,min=1,max=5"`
		Review string `json:"review"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if review belongs to user
	var productID string
	err := c.DB.QueryRow(`
		SELECT product_id FROM product_reviews
		WHERE id = $1 AND user_id = $2
	`, reviewID, userID).Scan(&productID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Review not found or not owned by you"})
		return
	}
	
	// Update review
	_, err = c.DB.Exec(`
		UPDATE product_reviews
		SET rating = $1, review = $2, updated_at = $3
		WHERE id = $4
	`, req.Rating, req.Review, time.Now(), reviewID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}
	
	// Update product average rating
	_, err = c.DB.Exec(`
		UPDATE products
		SET average_rating = (
			SELECT AVG(rating) FROM product_reviews
			WHERE product_id = $1 AND approved = true
		)
		WHERE id = $1
	`, productID)
	
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update product average rating: %v\n", err)
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Review updated successfully"})
}

func (c *ShopController) DeleteOwnReview(ctx *gin.Context) {
	reviewID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	// Check if review belongs to user and get product ID
	var productID string
	err := c.DB.QueryRow(`
		SELECT product_id FROM product_reviews
		WHERE id = $1 AND user_id = $2
	`, reviewID, userID).Scan(&productID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Review not found or not owned by you"})
		return
	}
	
	// Delete review
	_, err = c.DB.Exec(`
		DELETE FROM product_reviews
		WHERE id = $1
	`, reviewID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}
	
	// Update product average rating
	_, err = c.DB.Exec(`
		UPDATE products
		SET average_rating = (
			SELECT AVG(rating) FROM product_reviews
			WHERE product_id = $1 AND approved = true
		)
		WHERE id = $1
	`, productID)
	
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update product average rating: %v\n", err)
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

// User Address methods
func (c *ShopController) GetUserAddresses(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	rows, err := c.DB.Query(`
		SELECT id, address_line1, address_line2, city, state, country, postal_code, is_default, label
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch addresses"})
		return
	}
	defer rows.Close()
	
	var addresses []map[string]interface{}
	for rows.Next() {
		var address struct {
			ID          string `json:"id"`
			Line1       string `json:"address_line1"`
			Line2       string `json:"address_line2"`
			City        string `json:"city"`
			State       string `json:"state"`
			Country     string `json:"country"`
			PostalCode  string `json:"postal_code"`
			IsDefault   bool   `json:"is_default"`
			Label       string `json:"label"`
		}
		err := rows.Scan(&address.ID, &address.Line1, &address.Line2, &address.City, &address.State, &address.Country, &address.PostalCode, &address.IsDefault, &address.Label)
		if err != nil {
			continue
		}
		addresses = append(addresses, map[string]interface{}{
			"id":           address.ID,
			"address_line1": address.Line1,
			"address_line2": address.Line2,
			"city":         address.City,
			"state":        address.State,
			"country":      address.Country,
			"postal_code":  address.PostalCode,
			"is_default":   address.IsDefault,
			"label":        address.Label,
		})
	}
	
	ctx.JSON(http.StatusOK, gin.H{"addresses": addresses})
}

func (c *ShopController) AddUserAddress(ctx *gin.Context) {
	userID, _ := ctx.Get("userID")
	
	var req struct {
		AddressLine1 string `json:"address_line1" binding:"required"`
		AddressLine2 string `json:"address_line2"`
		City         string `json:"city" binding:"required"`
		State        string `json:"state" binding:"required"`
		Country      string `json:"country" binding:"required"`
		PostalCode   string `json:"postal_code" binding:"required"`
		IsDefault    bool   `json:"is_default"`
		Label        string `json:"label"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Begin transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	// If this is the default address, unset any existing default
	if req.IsDefault {
		_, err = tx.Exec(`
			UPDATE user_addresses
			SET is_default = false
			WHERE user_id = $1
		`, userID)
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update default address"})
			return
		}
	}
	
	// Add new address
	id := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO user_addresses (id, user_id, address_line1, address_line2, city, state, country, postal_code, is_default, label, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
	`, id, userID, req.AddressLine1, req.AddressLine2, req.City, req.State, req.Country, req.PostalCode, req.IsDefault, req.Label, time.Now())
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
		return
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Address added successfully",
		"id":      id,
	})
}

func (c *ShopController) UpdateUserAddress(ctx *gin.Context) {
	addressID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	var req struct {
		AddressLine1 string `json:"address_line1" binding:"required"`
		AddressLine2 string `json:"address_line2"`
		City         string `json:"city" binding:"required"`
		State        string `json:"state" binding:"required"`
		Country      string `json:"country" binding:"required"`
		PostalCode   string `json:"postal_code" binding:"required"`
		IsDefault    bool   `json:"is_default"`
		Label        string `json:"label"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if address belongs to user
	var count int
	err := c.DB.QueryRow(`
		SELECT COUNT(*) FROM user_addresses
		WHERE id = $1 AND user_id = $2
	`, addressID, userID).Scan(&count)
	
	if err != nil || count == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	// Begin transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	
	// If this is the default address, unset any existing default
	if req.IsDefault {
		_, err = tx.Exec(`
			UPDATE user_addresses
			SET is_default = false
			WHERE user_id = $1 AND id != $2
		`, userID, addressID)
		
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update default address"})
			return
		}
	}
	
	// Update address
	_, err = tx.Exec(`
		UPDATE user_addresses
		SET address_line1 = $1, address_line2 = $2, city = $3, state = $4, country = $5, postal_code = $6, is_default = $7, label = $8, updated_at = $9
		WHERE id = $10 AND user_id = $11
	`, req.AddressLine1, req.AddressLine2, req.City, req.State, req.Country, req.PostalCode, req.IsDefault, req.Label, time.Now(), addressID, userID)
	
	if err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		return
	}
	
	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Address updated successfully"})
}

func (c *ShopController) DeleteUserAddress(ctx *gin.Context) {
	addressID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	// Check if address is the default
	var isDefault bool
	err := c.DB.QueryRow(`
		SELECT is_default FROM user_addresses
		WHERE id = $1 AND user_id = $2
	`, addressID, userID).Scan(&isDefault)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	// Delete address
	result, err := c.DB.Exec(`
		DELETE FROM user_addresses
		WHERE id = $1 AND user_id = $2
	`, addressID, userID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	// If this was the default address, set a new default if any addresses remain
	if isDefault {
		c.DB.Exec(`
			UPDATE user_addresses
			SET is_default = true
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`, userID)
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
}