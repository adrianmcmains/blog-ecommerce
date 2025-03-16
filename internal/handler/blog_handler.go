package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/service"
	"github.com/gin-gonic/gin"
)

// BlogHandler handles blog-related requests
type BlogHandler struct {
	services *service.Service
	logger   *log.Logger
}

// NewBlogHandler creates a new blog handler
func NewBlogHandler(services *service.Service, logger *log.Logger) *BlogHandler {
	return &BlogHandler{
		services: services,
		logger:   logger,
	}
}

// GetAllPosts returns all blog posts with optional filtering
func (h *BlogHandler) GetAllPosts(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	// Get filter parameters
	category := c.Query("category")
	authorID := c.Query("author")
	
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Get posts from service
	posts, total, err := h.services.Blog.ListPosts(c.Request.Context(), page, limit, category, authorID)
	if err != nil {
		h.logger.Printf("Failed to get posts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get posts"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetPostByID returns a single post by ID
func (h *BlogHandler) GetPostByID(c *gin.Context) {
	id := c.Param("id")
	
	post, err := h.services.Blog.GetPost(c.Request.Context(), id)
	if err != nil {
		h.logger.Printf("Failed to get post: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	
	c.JSON(http.StatusOK, post)
}

// Convert models.PostInput to service.PostInput
func toServicePostInput(input models.PostInput) service.PostInput {
	return service.PostInput{
		Title:         input.Title,
		Content:       input.Content,
		Excerpt:       input.Excerpt,
		FeaturedImage: input.FeaturedImage,
		Categories:    input.Categories,
		Tags:          input.Tags,
		Published:     input.Published,
	}
}

// CreatePost creates a new blog post
func (h *BlogHandler) CreatePost(c *gin.Context) {
	var input models.PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid post input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	
	// Get user from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	
	// Convert to service.PostInput
	serviceInput := toServicePostInput(input)
	
	// Create post
	post, err := h.services.Blog.CreatePost(c.Request.Context(), serviceInput, userID.(string))
	if err != nil {
		h.logger.Printf("Failed to create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	
	c.JSON(http.StatusCreated, post)
}

// UpdatePost updates an existing blog post
func (h *BlogHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	
	var input models.PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid post input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	
	// Get user from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	
	// Convert to service.PostInput
	serviceInput := toServicePostInput(input)
	
	// Update post
	post, err := h.services.Blog.UpdatePost(c.Request.Context(), id, serviceInput, userID.(string))
	if err != nil {
		if err == service.ErrNotAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this post"})
			return
		}
		if err == service.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		
		h.logger.Printf("Failed to update post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}
	
	c.JSON(http.StatusOK, post)
}

// DeletePost deletes a blog post
func (h *BlogHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	
	// Get user from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	
	// Delete post
	err := h.services.Blog.DeletePost(c.Request.Context(), id, userID.(string))
	if err != nil {
		if err == service.ErrNotAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this post"})
			return
		}
		if err == service.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		
		h.logger.Printf("Failed to delete post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}