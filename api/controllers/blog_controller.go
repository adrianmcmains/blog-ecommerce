// File: api/controllers/blog_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/adrianmcmains/blog-ecommerce/api/blog"
)

// BlogController handles blog-related routes
type BlogController struct {
	DB *gorm.DB
}

// NewBlogController creates a new blog controller
func NewBlogController(db *gorm.DB) *BlogController {
	return &BlogController{DB: db}
}

// GetPosts retrieves a paginated list of posts
func (c *BlogController) GetPosts(ctx *gin.Context) {
	var params blog.PostQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults if not provided
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}

	offset := (params.Page - 1) * params.PageSize

	// Start building the query
	query := c.DB.Model(&blog.Post{})

	// Apply filters
	if params.CategorySlug != "" {
		var category blog.Category
		if err := c.DB.Where("slug = ?", params.CategorySlug).First(&category).Error; err == nil {
			query = query.Joins("JOIN post_categories ON post_categories.post_id = posts.id").
				Where("post_categories.category_id = ?", category.ID)
		}
	}

	if params.TagSlug != "" {
		var tag blog.Tag
		if err := c.DB.Where("slug = ?", params.TagSlug).First(&tag).Error; err == nil {
			query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
				Where("post_tags.tag_id = ?", tag.ID)
		}
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchTerm, searchTerm)
	}

	if params.Published != nil {
		query = query.Where("published = ?", *params.Published)
	}

	// Count total posts for pagination
	var total int64
	query.Count(&total)

	// Get the posts for the current page
	var posts []blog.Post
	query.Preload("Categories").Preload("Tags").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&posts)

	ctx.JSON(http.StatusOK, gin.H{
		"posts":      posts,
		"page":       params.Page,
		"page_size":  params.PageSize,
		"total":      total,
		"total_pages": (total + int64(params.PageSize) - 1) / int64(params.PageSize),
	})
}

// GetPost retrieves a single post by slug
func (c *BlogController) GetPost(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var post blog.Post
	if err := c.DB.Preload("Categories").Preload("Tags").Where("slug = ?", slug).First(&post).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// If post is not published, only allow author or admin to view it
	if !post.Published {
		userID, exists := ctx.Get("userID")
		if !exists {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		userRole, _ := ctx.Get("userRole")
		if post.AuthorID.String() != userID.(string) && userRole.(string) != "admin" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}
	}

	// Get author information
	var author blog.Author
	c.DB.Table("users").Select("id, first_name, last_name").Where("id = ?", post.AuthorID).Scan(&author)

	ctx.JSON(http.StatusOK, gin.H{
		"post":   post,
		"author": author,
	})
}

// CreatePost creates a new blog post
func (c *BlogController) CreatePost(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req blog.PostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug is unique
	var existingPost blog.Post
	if c.DB.Where("slug = ?", req.Slug).First(&existingPost).Error == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A post with this slug already exists"})
		return
	}

	authorID, err := uuid.Parse(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	now := time.Now()
	post := blog.Post{
		ID:           uuid.New(),
		Slug:         req.Slug,
		Title:        req.Title,
		Content:      req.Content,
		Excerpt:      req.Excerpt,
		FeaturedImage: req.FeaturedImage,
		AuthorID:     authorID,
		Published:    req.Published,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if req.Published {
		post.PublishedAt = &now
	}

	// Start a transaction to handle the post and its relationships
	err = c.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}

		// Handle categories
		if len(req.CategoryIDs) > 0 {
			var categories []blog.Category
			if err := tx.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
				return err
			}
			if err := tx.Model(&post).Association("Categories").Append(categories); err != nil {
				return err
			}
		}

		// Handle tags
		if len(req.TagIDs) > 0 {
			var tags []blog.Tag
			if err := tx.Where("id IN ?", req.TagIDs).Find(&tags).Error; err != nil {
				return err
			}
			if err := tx.Model(&post).Association("Tags").Append(tags); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create post: %v", err)})
		return
	}

	// Reload the post with associations
	c.DB.Preload("Categories").Preload("Tags").First(&post, post.ID)

	ctx.JSON(http.StatusCreated, gin.H{"post": post})
}

// UpdatePost updates an existing blog post
func (c *BlogController) UpdatePost(ctx *gin.Context) {
	slug := ctx.Param("slug")
	userID, exists := ctx.Get("userID")
	userRole, _ := ctx.Get("userRole")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var post blog.Post
	if err := c.DB.Where("slug = ?", slug).First(&post).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Only author or admin can update
	if post.AuthorID.String() != userID.(string) && userRole.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	var req blog.PostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new slug is different and unique
	if req.Slug != slug {
		var existingPost blog.Post
		if c.DB.Where("slug = ? AND id != ?", req.Slug, post.ID).First(&existingPost).Error == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "A post with this slug already exists"})
			return
		}
	}

	// Check if publishing status changed
	publishingNow := !post.Published && req.Published
	now := time.Now()

	// Update post fields
	post.Title = req.Title
	post.Slug = req.Slug
	post.Content = req.Content
	post.Excerpt = req.Excerpt
	post.FeaturedImage = req.FeaturedImage
	post.Published = req.Published
	post.UpdatedAt = now

	// If the post is being published for the first time
	if publishingNow {
		post.PublishedAt = &now
	}

	// Start a transaction to handle the post update and its relationships
	err := c.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&post).Error; err != nil {
			return err
		}

		// Update categories
		if err := tx.Model(&post).Association("Categories").Clear(); err != nil {
			return err
		}
		if len(req.CategoryIDs) > 0 {
			var categories []blog.Category
			if err := tx.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
				return err
			}
			if err := tx.Model(&post).Association("Categories").Append(categories); err != nil {
				return err
			}
		}

		// Update tags
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			return err
		}
		if len(req.TagIDs) > 0 {
			var tags []blog.Tag
			if err := tx.Where("id IN ?", req.TagIDs).Find(&tags).Error; err != nil {
				return err
			}
			if err := tx.Model(&post).Association("Tags").Append(tags); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update post: %v", err)})
		return
	}

	// Reload the post with associations
	c.DB.Preload("Categories").Preload("Tags").First(&post, post.ID)

	ctx.JSON(http.StatusOK, gin.H{"post": post})
}

// DeletePost deletes a blog post
func (c *BlogController) DeletePost(ctx *gin.Context) {
	slug := ctx.Param("slug")
	userID, exists := ctx.Get("userID")
	userRole, _ := ctx.Get("userRole")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var post blog.Post
	if err := c.DB.Where("slug = ?", slug).First(&post).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Only author or admin can delete
	if post.AuthorID.String() != userID.(string) && userRole.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	// Start a transaction to handle the deletion
	err := c.DB.Transaction(func(tx *gorm.DB) error {
		// Clear associations first
		if err := tx.Model(&post).Association("Categories").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
			return err
		}

		// Delete comments
		if err := tx.Where("post_id = ?", post.ID).Delete(&blog.Comment{}).Error; err != nil {
			return err
		}

		// Delete post
		if err := tx.Delete(&post).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete post: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// GetCategories retrieves all categories
func (c *BlogController) GetCategories(ctx *gin.Context) {
	var categories []blog.Category
	c.DB.Find(&categories)
	ctx.JSON(http.StatusOK, gin.H{"categories": categories})
}

// CreateCategory creates a new category
func (c *BlogController) CreateCategory(ctx *gin.Context) {
	var req blog.CategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug is unique
	var existingCategory blog.Category
	if c.DB.Where("slug = ?", req.Slug).First(&existingCategory).Error == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A category with this slug already exists"})
		return
	}

	category := blog.Category{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := c.DB.Create(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"category": category})
}

// UpdateCategory updates an existing category
func (c *BlogController) UpdateCategory(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var category blog.Category
	if err := c.DB.Where("slug = ?", slug).First(&category).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var req blog.CategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new slug is different and unique
	if req.Slug != slug {
		var existingCategory blog.Category
		if c.DB.Where("slug = ? AND id != ?", req.Slug, category.ID).First(&existingCategory).Error == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "A category with this slug already exists"})
			return
		}
	}

	// Update category fields
	category.Name = req.Name
	category.Slug = req.Slug
	category.Description = req.Description
	category.UpdatedAt = time.Now()

	if err := c.DB.Save(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"category": category})
}

// DeleteCategory deletes a category
func (c *BlogController) DeleteCategory(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var category blog.Category
	if err := c.DB.Where("slug = ?", slug).First(&category).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Check if category is in use
	var count int64
	c.DB.Model(&category).Association("Posts").Count()
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category as it is associated with posts"})
		return
	}

	if err := c.DB.Delete(&category).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// GetTags retrieves all tags
func (c *BlogController) GetTags(ctx *gin.Context) {
	var tags []blog.Tag
	c.DB.Find(&tags)
	ctx.JSON(http.StatusOK, gin.H{"tags": tags})
}

// CreateTag creates a new tag
func (c *BlogController) CreateTag(ctx *gin.Context) {
	var req blog.TagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug is unique
	var existingTag blog.Tag
	if c.DB.Where("slug = ?", req.Slug).First(&existingTag).Error == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A tag with this slug already exists"})
		return
	}

	tag := blog.Tag{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.DB.Create(&tag).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"tag": tag})
}

// UpdateTag updates an existing tag
func (c *BlogController) UpdateTag(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var tag blog.Tag
	if err := c.DB.Where("slug = ?", slug).First(&tag).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	var req blog.TagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new slug is different and unique
	if req.Slug != slug {
		var existingTag blog.Tag
		if c.DB.Where("slug = ? AND id != ?", req.Slug, tag.ID).First(&existingTag).Error == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "A tag with this slug already exists"})
			return
		}
	}

	// Update tag fields
	tag.Name = req.Name
	tag.Slug = req.Slug
	tag.UpdatedAt = time.Now()

	if err := c.DB.Save(&tag).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"tag": tag})
}

// DeleteTag deletes a tag
func (c *BlogController) DeleteTag(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var tag blog.Tag
	if err := c.DB.Where("slug = ?", slug).First(&tag).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	// Check if tag is in use
	var count int64
	c.DB.Model(&tag).Association("Posts").Count()
	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot delete tag as it is associated with posts"})
		return
	}

	if err := c.DB.Delete(&tag).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// AddComment adds a comment to a post
func (c *BlogController) AddComment(ctx *gin.Context) {
	slug := ctx.Param("slug")
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var post blog.Post
	if err := c.DB.Where("slug = ?", slug).First(&post).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var req blog.CommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commentUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Users with role admin or contributor have their comments auto-approved
	userRole, _ := ctx.Get("userRole")
	autoApprove := false
	if role, ok := userRole.(string); ok {
		autoApprove = (role == "admin" || role == "contributor")
	}

	comment := blog.Comment{
		ID:        uuid.New(),
		PostID:    post.ID,
		UserID:    commentUserID,
		Content:   req.Content,
		Approved:  autoApprove,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.DB.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"comment": comment})
}

// GetComments retrieves all comments for a post
func (c *BlogController) GetComments(ctx *gin.Context) {
	slug := ctx.Param("slug")

	var post blog.Post
	if err := c.DB.Where("slug = ?", slug).First(&post).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var comments []blog.Comment
	c.DB.Where("post_id = ?", post.ID).Order("created_at DESC").Find(&comments)

	// If user is not admin, only return approved comments
	userRole, exists := ctx.Get("userRole")
	if !exists || userRole.(string) != "admin" {
		var approvedComments []blog.Comment
		for _, comment := range comments {
			if comment.Approved {
				approvedComments = append(approvedComments, comment)
			}
		}
		comments = approvedComments
	}

	ctx.JSON(http.StatusOK, gin.H{"comments": comments})
}

// ApproveComment approves a comment (admin only)
func (c *BlogController) ApproveComment(ctx *gin.Context) {
	commentID := ctx.Param("id")

	var comment blog.Comment
	if err := c.DB.First(&comment, "id = ?", commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	comment.Approved = true
	comment.UpdatedAt = time.Now()

	if err := c.DB.Save(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve comment"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"comment": comment})
}

// DeleteComment deletes a comment
func (c *BlogController) DeleteComment(ctx *gin.Context) {
	commentID := ctx.Param("id")
	userID, exists := ctx.Get("userID")
	userRole, _ := ctx.Get("userRole")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var comment blog.Comment
	if err := c.DB.First(&comment, "id = ?", commentID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Only comment author or admin can delete
	if comment.UserID.String() != userID.(string) && userRole.(string) != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.DB.Delete(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}