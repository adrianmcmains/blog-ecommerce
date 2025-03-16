package handler

import (
	"log"

	"github.com/adrianmcmains/blog-ecommerce/internal/service"
)

// Handler wraps all the different handlers
type Handler struct {
    Auth *AuthHandler
    Blog *BlogHandler
    Shop *ShopHandler
}

// NewHandler creates a new handler instance
func NewHandler(services *service.Service, logger *log.Logger) *Handler {
    return &Handler{
        Auth: NewAuthHandler(services, logger),
		Blog: NewBlogHandler(services, logger),
		Shop: NewShopHandler(services, logger),
    }
}