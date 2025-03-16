package repository

import (
    "context"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type UsersRepo struct {
    db *gorm.DB
}

func NewUsersRepo(db *gorm.DB) *UsersRepo {
    return &UsersRepo{
        db: db,
    }
}

func (r *UsersRepo) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UsersRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
    var user models.User
    
    uuid, err := uuid.Parse(id)
    if err != nil {
        return nil, err
    }
    
    if err := r.db.WithContext(ctx).First(&user, uuid).Error; err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (r *UsersRepo) Update(ctx context.Context, id string, user *models.User) error {
    uuid, err := uuid.Parse(id)
    if err != nil {
        return err
    }
    
    return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", uuid).Updates(user).Error
}