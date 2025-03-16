package service

import (
    "context"
    "errors"
    //"time"
    
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/adrianmcmains/blog-ecommerce/internal/repository"
    "github.com/adrianmcmains/blog-ecommerce/pkg/util"
    
    "github.com/google/uuid"
)

type ShopService struct {
    productsRepo  repository.Products
    cartItemsRepo repository.CartItems
    ordersRepo    repository.Orders
}

func NewShopService(
    productsRepo repository.Products,
    cartItemsRepo repository.CartItems,
    ordersRepo repository.Orders,
) *ShopService {
    return &ShopService{
        productsRepo:  productsRepo,
        cartItemsRepo: cartItemsRepo,
        ordersRepo:    ordersRepo,
    }
}

// Product methods
func (s *ShopService) CreateProduct(ctx context.Context, input models.CreateProductInput) (*models.Product, error) {
    // Create product
    product := &models.Product{
        Name:          input.Name,
        Slug:          util.GenerateSlug(input.Name),
        Description:   input.Description,
        Price:         input.Price,
        StockQuantity: input.StockQuantity,
        Image:         input.Image,
        Status:        "active",
    }
    
    // Add categories
    if len(input.Categories) > 0 {
        for _, categoryName := range input.Categories {
            product.Categories = append(product.Categories, models.Category{
                Name: categoryName,
                Slug: util.GenerateSlug(categoryName),
            })
        }
    }
    
    // Add variants
    if len(input.Variants) > 0 {
        for _, variantInput := range input.Variants {
            product.Variants = append(product.Variants, models.ProductVariant{
                Name:          variantInput.Name,
                Price:         variantInput.Price,
                StockQuantity: variantInput.StockQuantity,
                SKU:           variantInput.SKU,
            })
        }
    }
    
    // Save product
    if err := s.productsRepo.Create(ctx, product); err != nil {
        return nil, err
    }
    
    return product, nil
}

func (s *ShopService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    return s.productsRepo.GetByID(ctx, id)
}

func (s *ShopService) ListProducts(ctx context.Context, filter models.ProductFilter) ([]models.Product, int64, error) {
    return s.productsRepo.List(ctx, filter)
}

func (s *ShopService) UpdateProduct(ctx context.Context, id string, input models.CreateProductInput) (*models.Product, error) {
    // Get existing product
    product, err := s.productsRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Update fields
    product.Name = input.Name
    product.Slug = util.GenerateSlug(input.Name)
    product.Description = input.Description
    product.Price = input.Price
    product.StockQuantity = input.StockQuantity
    product.Image = input.Image
    
    // Save product
    if err := s.productsRepo.Update(ctx, id, product); err != nil {
        return nil, err
    }
    
    return product, nil
}

func (s *ShopService) DeleteProduct(ctx context.Context, id string) error {
    return s.productsRepo.Delete(ctx, id)
}

// Cart methods
func (s *ShopService) AddToCart(ctx context.Context, input models.AddToCartInput) (*models.CartItem, error) {
    // Validate product
    productID, err := uuid.Parse(input.ProductID)
    if err != nil {
        return nil, err
    }
    
    product, err := s.productsRepo.GetByID(ctx, input.ProductID)
    if err != nil {
        return nil, errors.New("product not found")
    }
    
    // Check stock
    if product.StockQuantity < input.Quantity {
        return nil, errors.New("not enough stock")
    }
    
    // Parse user ID
    userID, err := uuid.Parse(input.UserID)
    if err != nil {
        return nil, err
    }
    
    // Create cart item
    cartItem := &models.CartItem{
        UserID:    userID,
        ProductID: productID,
        Quantity:  input.Quantity,
    }
    
    // Add variant if specified
    if input.VariantID != "" {
        variantID, err := uuid.Parse(input.VariantID)
        if err != nil {
            return nil, err
        }
        cartItem.VariantID = &variantID
    }
    
    // Save cart item
    if err := s.cartItemsRepo.Create(ctx, cartItem); err != nil {
        return nil, err
    }
    
    return cartItem, nil
}

func (s *ShopService) GetCart(ctx context.Context, userID string) ([]models.CartItem, error) {
    return s.cartItemsRepo.GetUserCart(ctx, userID)
}

func (s *ShopService) UpdateCartItem(ctx context.Context, id string, quantity int) (*models.CartItem, error) {
    // Get existing cart item
    cartItem, err := s.cartItemsRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Update quantity
    cartItem.Quantity = quantity
    
    // Save cart item
    if err := s.cartItemsRepo.Update(ctx, id, cartItem); err != nil {
        return nil, err
    }
    
    return cartItem, nil
}

func (s *ShopService) RemoveFromCart(ctx context.Context, id string) error {
    return s.cartItemsRepo.Delete(ctx, id)
}

// Order methods
func (s *ShopService) CreateOrder(ctx context.Context, input models.CreateOrderInput) (*models.Order, error) {
    // Parse user ID
    userID, err := uuid.Parse(input.UserID)
    if err != nil {
        return nil, err
    }
    
    // Get cart items
    var cartItems []models.CartItem
    var totalAmount float64
    
    for _, cartID := range input.CartIDs {
        cartItem, err := s.cartItemsRepo.GetByID(ctx, cartID)
        if err != nil {
            return nil, err
        }
        
        cartItems = append(cartItems, *cartItem)
        
        // Calculate price
        price := cartItem.Product.Price
        if cartItem.Variant != nil {
            price = cartItem.Variant.Price
        }
        
        totalAmount += price * float64(cartItem.Quantity)
    }
    
    // Create order
    order := &models.Order{
        UserID:      userID,
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
    
    // Save order
    if err := s.ordersRepo.Create(ctx, order); err != nil {
        return nil, err
    }
    
    // Clear cart
    for _, cartItem := range cartItems {
        if err := s.cartItemsRepo.Delete(ctx, cartItem.ID.String()); err != nil {
            // Log error but continue
            continue
        }
        
        // Update product stock
        if err := s.productsRepo.UpdateStock(ctx, cartItem.ProductID.String(), cartItem.Quantity); err != nil {
            // Log error but continue
            continue
        }
    }
    
    return order, nil
}

func (s *ShopService) GetOrder(ctx context.Context, id string, userID string) (*models.Order, error) {
    order, err := s.ordersRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Check if order belongs to user
    if order.UserID.String() != userID {
        return nil, errors.New("order not found")
    }
    
    return order, nil
}

func (s *ShopService) ListOrders(ctx context.Context, filter models.OrderFilter) ([]models.Order, int64, error) {
    return s.ordersRepo.List(ctx, filter)
}