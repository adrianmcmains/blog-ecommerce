package repository

import (
    "context"
    "strconv"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ProductsRepo struct {
    db *gorm.DB
}

func NewProductsRepo(db *gorm.DB) *ProductsRepo {
    return &ProductsRepo{
        db: db,
    }
}

func (r *ProductsRepo) Create(ctx context.Context, product *models.Product) error {
    return r.db.WithContext(ctx).Create(product).Error
}

func (r *ProductsRepo) GetByID(ctx context.Context, id string) (*models.Product, error) {
    var product models.Product
    
    uuid, err := uuid.Parse(id)
    if err != nil {
        return nil, err
    }
    
    if err := r.db.WithContext(ctx).
        Preload("Categories").
        Preload("Variants").
        First(&product, uuid).Error; err != nil {
        return nil, err
    }
    
    return &product, nil
}

func (r *ProductsRepo) List(ctx context.Context, filter models.ProductFilter) ([]models.Product, int64, error) {
    var products []models.Product
    var total int64
    
    query := r.db.WithContext(ctx).Model(&models.Product{})
    
    // Apply filters
    if filter.Category != "" {
        query = query.Joins("JOIN product_categories ON products.id = product_categories.product_id").
            Joins("JOIN categories ON product_categories.category_id = categories.id").
            Where("categories.slug = ?", filter.Category)
    }
    
    if filter.MinPrice != "" {
        minPrice, err := strconv.ParseFloat(filter.MinPrice, 64)
        if err == nil {
            query = query.Where("price >= ?", minPrice)
        }
    }
    
    if filter.MaxPrice != "" {
        maxPrice, err := strconv.ParseFloat(filter.MaxPrice, 64)
        if err == nil {
            query = query.Where("price <= ?", maxPrice)
        }
    }
    
    if filter.InStock != nil && *filter.InStock {
        query = query.Where("stock_quantity > 0")
    }
    
    // Count total records
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Pagination
    page, _ := strconv.Atoi(filter.Page)
    pageSize, _ := strconv.Atoi(filter.PageSize)
    
    if page < 1 {
        page = 1
    }
    
    if pageSize < 1 || pageSize > 100 {
        pageSize = 10
    }
    
    offset := (page - 1) * pageSize
    
    if err := query.
        Preload("Categories").
        Preload("Variants").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&products).Error; err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}

func (r *ProductsRepo) Update(ctx context.Context, id string, product *models.Product) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", uuid).Updates(product).Error
}

func (r *ProductsRepo) Delete(ctx context.Context, id string) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Delete(&models.Product{}, uuid).Error
}

func (r *ProductsRepo) UpdateStock(ctx context.Context, id string, quantity int) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).
        Model(&models.Product{}).
        Where("id = ?", uuid).
        Update("stock_quantity", gorm.Expr("stock_quantity - ?", quantity)).
        Error
}