// api/controllers/shop_controller.go
package controllers

import (
	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	//"time"

	"github.com/gin-gonic/gin"
)

// ShopController handles shop-related requests
type ShopController struct {
	db *sql.DB
}

// NewShopController creates a new shop controller
func NewShopController(db *sql.DB) *ShopController {
	return &ShopController{db: db}
}

// GetProducts gets all products
func (c *ShopController) GetProducts(ctx *gin.Context) {
	// Extract query parameters
	categorySlug := ctx.Query("category")
	searchQuery := ctx.Query("query")
	featuredOnly := ctx.Query("featured") == "true"
	sortBy := ctx.DefaultQuery("sort", "id")
	sortOrder := ctx.DefaultQuery("order", "asc")
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	// Convert limit and offset to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build query
	query := `
		SELECT 
			p.id, p.name, p.slug, p.description, p.price, p.sale_price,
			p.stock, p.sku, p.featured, p.visible, p.created_at, p.updated_at
		FROM products p
		WHERE p.visible = true
	`

	// Add filters
	args := []interface{}{}
	if categorySlug != "" {
		query += `
			AND p.id IN (
				SELECT pc.product_id
				FROM product_category_items pc
				JOIN product_categories c ON pc.category_id = c.id
				WHERE c.slug = $1
			)
		`
		args = append(args, categorySlug)
	}

	if searchQuery != "" {
		query += `
			AND (
				p.name ILIKE $%d OR
				p.description ILIKE $%d
			)
		`
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
		// Update placeholders
		query = fmt.Sprintf(query, len(args)-1, len(args))
	}

	if featuredOnly {
		query += " AND p.featured = true"
	}

	// Add sorting
	query += " ORDER BY p." + sanitizeSortColumn(sortBy) + " " + sanitizeSortOrder(sortOrder)

	// Add pagination
	query += " LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	// Execute query
	rows, err := c.db.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer rows.Close()

	// Process results
	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.SalePrice,
			&p.Stock, &p.SKU, &p.Featured, &p.Visible, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process products"})
			return
		}

		// Get product categories
		p.Categories, err = c.getProductCategories(p.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product categories"})
			return
		}

		// Get product images
		p.Images, err = c.getProductImages(p.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product images"})
			return
		}

		products = append(products, p)
	}

	// Count total products (for pagination)
	var totalCount int
	countQuery := strings.Replace(query, 
		"SELECT \n\t\t\tp.id, p.name, p.slug, p.description, p.price, p.sale_price,\n\t\t\tp.stock, p.sku, p.featured, p.visible, p.created_at, p.updated_at", 
		"SELECT COUNT(*)", 
		1)
	countQuery = countQuery[:strings.LastIndex(countQuery, "ORDER BY")]
	err = c.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"products": products,
		"total": totalCount,
		"limit": limit,
		"offset": offset,
	})
}

// GetProductBySlug gets a product by its slug
func (c *ShopController) GetProductBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	// Get product details
	var p models.Product
	err := c.db.QueryRow(`
		SELECT 
			id, name, slug, description, price, sale_price,
			stock, sku, featured, visible, created_at, updated_at
		FROM products
		WHERE slug = $1 AND visible = true
	`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.SalePrice,
		&p.Stock, &p.SKU, &p.Featured, &p.Visible, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	// Get product categories
	p.Categories, err = c.getProductCategories(p.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product categories"})
		return
	}

	// Get product images
	p.Images, err = c.getProductImages(p.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product images"})
		return
	}

	ctx.JSON(http.StatusOK, p)
}

// GetCategories gets all product categories
func (c *ShopController) GetCategories(ctx *gin.Context) {
	rows, err := c.db.Query(`
		SELECT id, name, slug, description, image, created_at, updated_at
		FROM product_categories
		ORDER BY name
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	defer rows.Close()

	var categories []models.ProductCategory
	for rows.Next() {
		var cat models.ProductCategory
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.Image, &cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process categories"})
			return
		}
		categories = append(categories, cat)
	}

	ctx.JSON(http.StatusOK, categories)
}

// Helper functions
func (c *ShopController) getProductCategories(productID int) ([]string, error) {
	rows, err := c.db.Query(`
		SELECT c.name
		FROM product_categories c
		JOIN product_category_items ci ON c.id = ci.category_id
		WHERE ci.product_id = $1
		ORDER BY c.name
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (c *ShopController) getProductImages(productID int) ([]string, error) {
	rows, err := c.db.Query(`
		SELECT image_url
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var image string
		if err := rows.Scan(&image); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

// Sanitize sort column to prevent SQL injection
func sanitizeSortColumn(column string) string {
	allowedColumns := map[string]bool{
		"id": true, "name": true, "price": true, "created_at": true, "updated_at": true,
	}
	if allowedColumns[column] {
		return column
	}
	return "id"
}

// Sanitize sort order to prevent SQL injection
func sanitizeSortOrder(order string) string {
	if strings.ToLower(order) == "desc" {
		return "DESC"
	}
	return "ASC"
}

// Admin product management endpoints
func (c *ShopController) CreateProduct(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Insert product
	err = tx.QueryRow(`
		INSERT INTO products (
			name, slug, description, price, sale_price, stock, sku, featured, visible, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`,
		product.Name, product.Slug, product.Description, product.Price, product.SalePrice,
		product.Stock, product.SKU, product.Featured, product.Visible,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Add categories if provided
	if len(product.Categories) > 0 {
		for _, categoryName := range product.Categories {
			// Get category ID
			var categoryID int
			err = tx.QueryRow("SELECT id FROM product_categories WHERE name = $1", categoryName).Scan(&categoryID)
			if err != nil {
				// Category doesn't exist, create it
				if err == sql.ErrNoRows {
					slug := strings.ToLower(strings.ReplaceAll(categoryName, " ", "-"))
					err = tx.QueryRow(`
						INSERT INTO product_categories (name, slug, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())
						RETURNING id
					`, categoryName, slug).Scan(&categoryID)
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
						return
					}
				} else {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category"})
					return
				}
			}

			// Link product to category
			_, err = tx.Exec(`
				INSERT INTO product_category_items (product_id, category_id)
				VALUES ($1, $2)
			`, product.ID, categoryID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link product to category"})
				return
			}
		}
	}

	// Add images if provided
	if len(product.Images) > 0 {
		for i, imageURL := range product.Images {
			isPrimary := i == 0 // First image is primary
			_, err = tx.Exec(`
				INSERT INTO product_images (product_id, image_url, is_primary, sort_order, created_at)
				VALUES ($1, $2, $3, $4, NOW())
			`, product.ID, imageURL, isPrimary, i)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product image"})
				return
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	ctx.JSON(http.StatusCreated, product)
}

func (c *ShopController) UpdateProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Update product
	_, err = tx.Exec(`
		UPDATE products
		SET name = $1, slug = $2, description = $3, price = $4, sale_price = $5,
			stock = $6, sku = $7, featured = $8, visible = $9, updated_at = NOW()
		WHERE id = $10
	`,
		product.Name, product.Slug, product.Description, product.Price, product.SalePrice,
		product.Stock, product.SKU, product.Featured, product.Visible, productID,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Update categories if provided
	if len(product.Categories) > 0 {
		// Remove existing category links
		_, err = tx.Exec("DELETE FROM product_category_items WHERE product_id = $1", productID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product categories"})
			return
		}

		// Add new category links
		for _, categoryName := range product.Categories {
			// Get category ID
			var categoryID int
			err = tx.QueryRow("SELECT id FROM product_categories WHERE name = $1", categoryName).Scan(&categoryID)
			if err != nil {
				// Category doesn't exist, create it
				if err == sql.ErrNoRows {
					slug := strings.ToLower(strings.ReplaceAll(categoryName, " ", "-"))
					err = tx.QueryRow(`
						INSERT INTO product_categories (name, slug, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())
						RETURNING id
					`, categoryName, slug).Scan(&categoryID)
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
						return
					}
				} else {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category"})
					return
				}
			}

			// Link product to category
			_, err = tx.Exec(`
				INSERT INTO product_category_items (product_id, category_id)
				VALUES ($1, $2)
			`, productID, categoryID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link product to category"})
				return
			}
		}
	}

	// Update images if provided
	if len(product.Images) > 0 {
		// Remove existing images
		_, err = tx.Exec("DELETE FROM product_images WHERE product_id = $1", productID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product images"})
			return
		}

		// Add new images
		for i, imageURL := range product.Images {
			isPrimary := i == 0 // First image is primary
			_, err = tx.Exec(`
				INSERT INTO product_images (product_id, image_url, is_primary, sort_order, created_at)
				VALUES ($1, $2, $3, $4, NOW())
			`, productID, imageURL, isPrimary, i)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product image"})
				return
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated product
	product.ID = productID
	product, err = c.getProductByID(productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated product"})
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (c *ShopController) DeleteProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Delete product category links
	_, err = tx.Exec("DELETE FROM product_category_items WHERE product_id = $1", productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product category links"})
		return
	}

	// Delete product images
	_, err = tx.Exec("DELETE FROM product_images WHERE product_id = $1", productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product images"})
		return
	}

	// Delete product
	result, err := tx.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	// Check if product was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
		return
	}
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// Category management endpoints
func (c *ShopController) CreateCategory(ctx *gin.Context) {
	var category models.ProductCategory
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert category
	err := c.db.QueryRow(`
		INSERT INTO product_categories (name, slug, description, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, category.Name, category.Slug, category.Description, category.Image).Scan(
		&category.ID, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	ctx.JSON(http.StatusCreated, category)
}

func (c *ShopController) UpdateCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	categoryID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.ProductCategory
	if err := ctx.ShouldBindJSON(&category); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update category
	_, err = c.db.Exec(`
		UPDATE product_categories
		SET name = $1, slug = $2, description = $3, image = $4, updated_at = NOW()
		WHERE id = $5
	`, category.Name, category.Slug, category.Description, category.Image, categoryID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	// Get updated category
	category.ID = categoryID
	err = c.db.QueryRow(`
		SELECT created_at, updated_at
		FROM product_categories
		WHERE id = $1
	`, categoryID).Scan(&category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated category"})
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (c *ShopController) DeleteCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	categoryID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Delete category links
	_, err = tx.Exec("DELETE FROM product_category_items WHERE category_id = $1", categoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category links"})
		return
	}

	// Delete category
	result, err := tx.Exec("DELETE FROM product_categories WHERE id = $1", categoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	// Check if category was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
		return
	}
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// Helper to get product by ID
func (c *ShopController) getProductByID(productID int) (models.Product, error) {
	var p models.Product
	err := c.db.QueryRow(`
		SELECT 
			id, name, slug, description, price, sale_price,
			stock, sku, featured, visible, created_at, updated_at
		FROM products
		WHERE id = $1
	`, productID).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.SalePrice,
		&p.Stock, &p.SKU, &p.Featured, &p.Visible, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return p, err
	}

	// Get product categories
	p.Categories, err = c.getProductCategories(p.ID)
	if err != nil {
		return p, err
	}

	// Get product images
	p.Images, err = c.getProductImages(p.ID)
	if err != nil {
		return p, err
	}

	return p, nil
}