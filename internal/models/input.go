package models

// Auth inputs
type RegisterInput struct {
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=6"`
    FirstName string `json:"first_name" binding:"required"`
    LastName  string `json:"last_name" binding:"required"`
    Role      string `json:"role" binding:"required,oneof=admin customer contributor"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// Blog inputs
type CreatePostInput struct {
    Title      string   `json:"title" binding:"required"`
    Content    string   `json:"content" binding:"required"`
    AuthorID   string   `json:"-"`
    Image      string   `json:"image"`
    Categories []string `json:"categories"`
    Tags       []string `json:"tags"`
}

type UpdatePostInput struct {
    Title      *string   `json:"title"`
    Content    *string   `json:"content"`
    Image      *string   `json:"image"`
    Categories *[]string `json:"categories"`
    Tags       *[]string `json:"tags"`
}

type PostFilter struct {
    Page     string `form:"page,default=1"`
    PageSize string `form:"page_size,default=10"`
    Category string `form:"category"`
    Tag      string `form:"tag"`
    AuthorID string `form:"author_id"`
}

// Shop inputs
type CreateProductInput struct {
    Name          string         `json:"name" binding:"required"`
    Description   string         `json:"description" binding:"required"`
    Price         float64        `json:"price" binding:"required,gt=0"`
    StockQuantity int           `json:"stock_quantity" binding:"required,gte=0"`
    Image         string         `json:"image"`
    Categories    []string       `json:"categories"`
    Variants      []VariantInput `json:"variants"`
}

type VariantInput struct {
    Name          string  `json:"name" binding:"required"`
    Price         float64 `json:"price" binding:"required,gt=0"`
    StockQuantity int    `json:"stock_quantity" binding:"required,gte=0"`
    SKU           string  `json:"sku" binding:"required"`
}

type ProductFilter struct {
    Page      string  `form:"page,default=1"`
    PageSize  string  `form:"page_size,default=10"`
    Category  string  `form:"category"`
    MinPrice  string  `form:"min_price"`
    MaxPrice  string  `form:"max_price"`
    InStock   *bool   `form:"in_stock"`
}

type AddToCartInput struct {
    ProductID  string `json:"product_id" binding:"required"`
    VariantID  string `json:"variant_id"`
    Quantity   int    `json:"quantity" binding:"required,gt=0"`
    UserID     string `json:"-"`
}

type CreateOrderInput struct {
    UserID   string        `json:"-"`
    Address  AddressInput  `json:"address" binding:"required"`
    CartIDs  []string     `json:"cart_ids" binding:"required"`
}

type AddressInput struct {
    Street     string `json:"street" binding:"required"`
    City       string `json:"city" binding:"required"`
    State      string `json:"state" binding:"required"`
    Country    string `json:"country" binding:"required"`
    PostalCode string `json:"postal_code" binding:"required"`
}

type OrderFilter struct {
    UserID    string `form:"-"`
    Page      string `form:"page,default=1"`
    PageSize  string `form:"page_size,default=10"`
    Status    string `form:"status"`
}

// In models/input.go
type RefreshTokenInput struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}