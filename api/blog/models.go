// File: api/blog/models.go
package blog

import (
	"time"

	"github.com/google/uuid"
	//"gorm.io/gorm"
)

// Post represents a blog post
type Post struct {
	ID           uuid.UUID   `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Slug         string      `json:"slug" gorm:"uniqueIndex;not null"`
	Title        string      `json:"title" gorm:"not null"`
	Content      string      `json:"content" gorm:"type:text;not null"`
	Excerpt      string      `json:"excerpt" gorm:"type:text"`
	FeaturedImage string      `json:"featured_image"`
	AuthorID     uuid.UUID   `json:"author_id" gorm:"type:uuid;not null"`
	Published    bool        `json:"published" gorm:"default:false"`
	PublishedAt  *time.Time  `json:"published_at"`
	CreatedAt    time.Time   `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"not null"`
	Categories   []Category  `json:"categories" gorm:"many2many:post_categories;"`
	Tags         []Tag       `json:"tags" gorm:"many2many:post_tags;"`
	Comments     []Comment   `json:"comments" gorm:"foreignKey:PostID"`
}

// Category represents a blog category
type Category struct {
	ID          uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`
	Posts       []Post    `json:"posts" gorm:"many2many:post_categories;"`
}

// Tag represents a blog tag
type Tag struct {
	ID        uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
	Posts     []Post    `json:"posts" gorm:"many2many:post_tags;"`
}

// Comment represents a comment on a blog post
type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Approved  bool      `json:"approved" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// PostRequest represents the form data for creating/updating a post
type PostRequest struct {
	Title        string    `json:"title" binding:"required"`
	Slug         string    `json:"slug" binding:"required"`
	Content      string    `json:"content" binding:"required"`
	Excerpt      string    `json:"excerpt"`
	FeaturedImage string    `json:"featured_image"`
	Published    bool      `json:"published"`
	CategoryIDs  []string  `json:"category_ids"`
	TagIDs       []string  `json:"tag_ids"`
}

// PostResponse is the response containing post data
type PostResponse struct {
	Post       Post       `json:"post"`
	Categories []Category `json:"categories"`
	Tags       []Tag      `json:"tags"`
	Author     Author     `json:"author"`
}

// Author represents a simplified user object for post authorship
type Author struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// CategoryRequest is the request for creating/updating a category
type CategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

// TagRequest is the request for creating/updating a tag
type TagRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// CommentRequest is the request for creating a comment
type CommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// PaginationParams contains pagination parameters
type PaginationParams struct {
	Page     int `form:"page,default=1" json:"page"`
	PageSize int `form:"page_size,default=10" json:"page_size"`
}

// PostQueryParams contains filters for post queries
type PostQueryParams struct {
	PaginationParams
	CategorySlug string `form:"category" json:"category"`
	TagSlug      string `form:"tag" json:"tag"`
	Search       string `form:"search" json:"search"`
	Published    *bool  `form:"published" json:"published"`
}