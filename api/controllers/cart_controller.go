// api/controllers/cart_controller.go
package controllers

import (
	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"database/sql"
	"net/http"
	"strconv"
	//"time"

	"github.com/gin-gonic/gin"
)

// Cart controller
func (c *ShopController) GetCart(ctx *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get or create cart
	cart, err := c.getOrCreateCart(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		return
	}

	ctx.JSON(http.StatusOK, cart)

}

// AddToCart adds a product to the cart
func (c *ShopController) AddToCart(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req struct {
		ProductID int `json:"productId" binding:"required"`
		Quantity  int `json:"quantity" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product exists and is in stock
	var product models.Product
	err := c.DB.QueryRow(`
		SELECT id, name, price, sale_price, stock
		FROM products
		WHERE id = $1 AND visible = true
	`, req.ProductID).Scan(&product.ID, &product.Name, &product.Price, &product.SalePrice, &product.Stock)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	// Check if product is in stock
	if product.Stock < req.Quantity {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock available"})
		return
	}

	// Get first product image
	var productImage string
	err = c.DB.QueryRow(`
		SELECT image_url
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order ASC
		LIMIT 1
	`, product.ID).Scan(&productImage)

	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product image"})
		return
	}

	// Start transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get or create cart
	var cartID int
	err = tx.QueryRow(`
		SELECT id
		FROM carts
		WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create new cart
			err = tx.QueryRow(`
				INSERT INTO carts (user_id, created_at, updated_at)
				VALUES ($1, NOW(), NOW())
				RETURNING id
			`, userID).Scan(&cartID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
			return
		}
	}

	// Check if product is already in cart
	var cartItemID int
	var currentQuantity int
	err = tx.QueryRow(`
		SELECT id, quantity
		FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, req.ProductID).Scan(&cartItemID, &currentQuantity)

	// Use sale price if available, otherwise use regular price
	productPrice := product.Price
	if product.SalePrice > 0 {
		productPrice = product.SalePrice
	}

	if err != nil {
		if err == sql.ErrNoRows {
			// Add new cart item
			_, err = tx.Exec(`
				INSERT INTO cart_items (
					cart_id, product_id, name, price, quantity, image, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`, cartID, product.ID, product.Name, productPrice, req.Quantity, productImage)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check cart"})
			return
		}
	} else {
		// Update existing cart item
		_, err = tx.Exec(`
			UPDATE cart_items
			SET quantity = $1, updated_at = NOW()
			WHERE id = $2
		`, currentQuantity+req.Quantity, cartItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	}

	// Update cart timestamp
	_, err = tx.Exec(`
		UPDATE carts
		SET updated_at = NOW()
		WHERE id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated cart
	cart, err := c.getOrCreateCart(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated cart"})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// UpdateCartItem updates a cart item quantity
func (c *ShopController) UpdateCartItem(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get cart item ID from URL
	itemID := ctx.Param("id")
	cartItemID, err := strconv.Atoi(itemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	// Parse request body
	var req struct {
		Quantity int `json:"quantity" binding:"required,min=0"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get cart ID for this user
	var cartID int
	err = tx.QueryRow(`
		SELECT id
		FROM carts
		WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		}
		return
	}

	// Verify cart item belongs to this user's cart
	var productID int
	err = tx.QueryRow(`
		SELECT product_id
		FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`, cartItemID, cartID).Scan(&productID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify cart item"})
		}
		return
	}

	// If quantity is 0, remove item
	if req.Quantity == 0 {
		_, err = tx.Exec(`
			DELETE FROM cart_items
			WHERE id = $1
		`, cartItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove cart item"})
			return
		}
	} else {
		// Check if product has enough stock
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

		if inStock < req.Quantity {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock available"})
			return
		}

		// Update cart item quantity
		_, err = tx.Exec(`
			UPDATE cart_items
			SET quantity = $1, updated_at = NOW()
			WHERE id = $2
		`, req.Quantity, cartItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	}

	// Update cart timestamp
	_, err = tx.Exec(`
		UPDATE carts
		SET updated_at = NOW()
		WHERE id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated cart
	cart, err := c.getOrCreateCart(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated cart"})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// RemoveFromCart removes an item from the cart
func (c *ShopController) RemoveFromCart(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get cart item ID from URL
	itemID := ctx.Param("id")
	cartItemID, err := strconv.Atoi(itemID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
		return
	}

	// Start transaction
	tx, err := c.DB.Begin()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get cart ID for this user
	var cartID int
	err = tx.QueryRow(`
		SELECT id
		FROM carts
		WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
		}
		return
	}

	// Verify and delete cart item
	result, err := tx.Exec(`
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`, cartItemID, cartID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove cart item"})
		return
	}

	// Check if item was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
		return
	}
	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// Update cart timestamp
	_, err = tx.Exec(`
		UPDATE carts
		SET updated_at = NOW()
		WHERE id = $1
	`, cartID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated cart
	cart, err := c.getOrCreateCart(userID.(int))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated cart"})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// getOrCreateCart gets the user's cart or creates a new one
func (c *ShopController) getOrCreateCart(userID int) (models.Cart, error) {
	var cart models.Cart
	cart.UserID = userID

	// Get cart ID for this user or create a new cart
	err := c.DB.QueryRow(`
		SELECT id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
	`, userID).Scan(&cart.ID, &cart.CreatedAt, &cart.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create new cart
			err = c.DB.QueryRow(`
				INSERT INTO carts (user_id, created_at, updated_at)
				VALUES ($1, NOW(), NOW())
				RETURNING id, created_at, updated_at
			`, userID).Scan(&cart.ID, &cart.CreatedAt, &cart.UpdatedAt)
			if err != nil {
				return cart, err
			}
		} else {
			return cart, err
		}
	}

	// Get cart items
	rows, err := c.DB.Query(`
		SELECT id, product_id, name, price, quantity, image, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at
	`, cart.ID)
	if err != nil {
		return cart, err
	}
	defer rows.Close()

	// Process cart items
	cart.Items = []models.CartItem{}
	cart.Total = 0
	for rows.Next() {
		var item models.CartItem
		err := rows.Scan(
			&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity,
			&item.Image, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return cart, err
		}

		item.CartID = cart.ID
		item.Total = item.Price * float64(item.Quantity)
		cart.Items = append(cart.Items, item)
		cart.Total += item.Total
	}

	return cart, nil
}