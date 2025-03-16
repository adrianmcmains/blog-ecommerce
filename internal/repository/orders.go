package repository

import (
    "context"
    "strconv"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type OrdersRepo struct {
    db *gorm.DB
}

func NewOrdersRepo(db *gorm.DB) *OrdersRepo {
    return &OrdersRepo{
        db: db,
    }
}

func (r *OrdersRepo) Create(ctx context.Context, order *models.Order) error {
    return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrdersRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
    var order models.Order
    
    uuid, err := uuid.Parse(id)
    if err != nil {
        return nil, err
    }
    
    if err := r.db.WithContext(ctx).
        Preload("User").
        Preload("Items").
        Preload("Items.Product").
        Preload("Items.Variant").
        First(&order, uuid).Error; err != nil {
        return nil, err
    }
    
    return &order, nil
}

func (r *OrdersRepo) List(ctx context.Context, filter models.OrderFilter) ([]models.Order, int64, error) {
    var orders []models.Order
    var total int64
    
    query := r.db.WithContext(ctx).Model(&models.Order{})
    
    // Apply filters
    userUUID, err := uuid.Parse(filter.UserID)
    if err == nil {
        query = query.Where("user_id = ?", userUUID)
    }
    
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
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
        Preload("User").
        Preload("Items").
        Preload("Items.Product").
        Preload("Items.Variant").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&orders).Error; err != nil {
        return nil, 0, err
    }
    
    return orders, total, nil
}

func (r *OrdersRepo) Update(ctx context.Context, id string, order *models.Order) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", uuid).Updates(order).Error
}