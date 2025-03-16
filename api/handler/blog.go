package handler

import (
	"net/http"
	"strconv"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/service"

	"github.com/gin-gonic/gin"
)

// BlogHandler handles blog-related requests
type BlogHandler struct {
	services *service.Service
}

// NewBlogHandler creates a new BlogHandler instance
func NewBlogHandler(services *service.Service) *BlogHandler {
	return &BlogHandler{
		services: services,
	}
}

// CreatePost handles creating a new blog post
func (h *BlogHandler) CreatePost(c *gin.Context) {
	var input models.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Convert to service.PostInput
	serviceInput := service.PostInput{
		Title:      input.Title,
		Content:    input.Content,
		Categories: input.Categories,
		Tags:       input.Tags,
	}

	// Set optional fields if they exist in your input model
	// These conditionals are assuming the fields are pointers in your model
	if input.Image != "" {
		serviceInput.FeaturedImage = input.Image
	}

	// Create post
	post, err := h.services.Blog.CreatePost(c.Request.Context(), serviceInput, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetPost handles retrieving a single blog post by ID
func (h *BlogHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	post, err := h.services.Blog.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// ListPosts handles retrieving a list of blog posts with pagination and filtering
func (h *BlogHandler) ListPosts(c *gin.Context) {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	
	// Get filter parameters
	category := c.Query("category")
	authorID := c.Query("author")

	// Call service with correct parameters
	posts, total, err := h.services.Blog.ListPosts(c.Request.Context(), page, limit, category, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"posts":      posts,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

// UpdatePost handles updating an existing blog post
func (h *BlogHandler) UpdatePost(c *gin.Context) {
    id := c.Param("id")
    
    var input models.UpdatePostInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
        return
    }
    userIDStr, ok := userID.(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
        return
    }

    // Create a service.PostInput with fields from the UpdatePostInput
    serviceInput := service.PostInput{}
    
    // Handle pointer fields correctly
    if input.Title != nil {
        serviceInput.Title = *input.Title
    }
    
    if input.Content != nil {
        serviceInput.Content = *input.Content
    }
    
    if input.Image != nil {
        serviceInput.FeaturedImage = *input.Image
    }
    
    if input.Categories != nil {
        serviceInput.Categories = *input.Categories
    }
    
    if input.Tags != nil {
        serviceInput.Tags = *input.Tags
    }
    
    // Update post
    post, err := h.services.Blog.UpdatePost(c.Request.Context(), id, serviceInput, userIDStr)
    if err != nil {
        switch err {
        case service.ErrPostNotFound:
            c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        case service.ErrNotAuthorized:
            c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this post"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, post)
}

// DeletePost handles deleting a blog post
func (h *BlogHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}
	
	// Delete post
	err := h.services.Blog.DeletePost(c.Request.Context(), id, userIDStr)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case service.ErrNotAuthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this post"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}