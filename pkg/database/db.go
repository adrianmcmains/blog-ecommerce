package database

import (
    //"log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
)

func InitDB(databaseURL string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Auto migrate the schemas
    err = db.AutoMigrate(
        &models.User{},
        &models.Post{},
        &models.Category{},
        &models.Tag{},
        &models.Product{},
        &models.ProductVariant{},
        &models.CartItem{},
        &models.Order{},
        &models.OrderItem{},
        &models.Payment{},
        &models.PaymentMethod{},
    )
    if err != nil {
        return nil, err
    }

    return db, nil
}