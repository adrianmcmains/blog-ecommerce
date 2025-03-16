package repository

import (
    "context"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type CartItemsRepo struct {
    db *gorm.DB
}

func NewCartItemsRepo(db *gorm.DB) *CartItemsRepo {
    return &CartItemsRepo{
        db: db,
    }
}

func (r *CartItemsRepo) Create(ctx context.Context, item *models.CartItem) error {
    return r.db.WithContext(ctx).Create(item).Error
}

func (r *CartItemsRepo) GetByID(ctx context.Context, id string) (*models.CartItem, error) {
    var item models.CartItem
    
    uuid, err := uuid.Parse(id)
    if err != nil {
        return nil, err
    }
    
    if err := r.db.WithContext(ctx).
        Preload("Product").
        Preload("Variant").
        First(&item, uuid).Error; err != nil {
        return nil, err
    }
    
    return &item, nil
}

func (r *CartItemsRepo) GetUserCart(ctx context.Context, userID string) ([]models.CartItem, error) {
    var items []models.CartItem
    
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        return nil, err
    }
    
    if err := r.db.WithContext(ctx).
        Where("user_id = ?", userUUID).
        Preload("Product").
        Preload("Variant").
        Find(&items).Error; err != nil {
        return nil, err
    }
    
    return items, nil
}

func (r *CartItemsRepo) Update(ctx context.Context, id string, item *models.CartItem) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Model(&models.CartItem{}).Where("id = ?", uuid).Updates(item).Error
}

func (r *CartItemsRepo) Delete(ctx context.Context, id string) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Delete(&models.CartItem{}, uuid).Error
}

func (r *CartItemsRepo) DeleteUserCart(ctx context.Context, userID string) error {
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Where("user_id = ?", userUUID).Delete(&models.CartItem{}).Error
}