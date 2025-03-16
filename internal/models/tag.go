package models

import (
	"github.com/google/uuid"
	"time"
)

// Tag represents a tag that can be applied to posts or products
type Tag struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"unique;not null" json:"slug"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	Posts     []Post    `gorm:"many2many:post_tags;" json:"posts,omitempty"`
	Products  []Product `gorm:"many2many:product_tags;" json:"products,omitempty"`
}

// TagInput represents the input for creating or updating a tag
type TagInput struct {
	Name string `json:"name" binding:"required"`
}

// TagList represents a list of tags with pagination metadata
type TagList struct {
	Tags  []Tag `json:"tags"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int   `json:"pages"`
}

// TagFilter represents filtering options for tag listing
type TagFilter struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	SortBy string `json:"sort_by"`
}

// PostTagInput represents the input for associating tags with a post
type PostTagInput struct {
	PostID string   `json:"post_id" binding:"required"`
	TagIDs []string `json:"tag_ids" binding:"required"`
}

// ProductTagInput represents the input for associating tags with a product
type ProductTagInput struct {
	ProductID string   `json:"product_id" binding:"required"`
	TagIDs    []string `json:"tag_ids" binding:"required"`
}

// TagResponse represents a tag with additional metadata
type TagResponse struct {
	Tag
	PostCount    int `json:"post_count"`
	ProductCount int `json:"product_count"`
}

// ValidateTag validates the tag input
func (t *TagInput) ValidateTag() map[string]string {
	errors := make(map[string]string)
	
	if t.Name == "" {
		errors["name"] = "Name is required"
	}
	
	return errors
}