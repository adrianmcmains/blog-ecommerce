package handler

import (
    "github.com/adrianmcmains/blog-ecommerce/internal/service"
)

type Handler struct {
    Auth *AuthHandler
    Blog *BlogHandler
    Shop *ShopHandler
}

func NewHandler(services *service.Service) *Handler {
    return &Handler{
        Auth: NewAuthHandler(services),
        Blog: NewBlogHandler(services),
        Shop: NewShopHandler(services),
    }
}