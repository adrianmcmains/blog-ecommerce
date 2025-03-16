package models

import (
	"time"
	// "github.com/google/uuid"
)

// Post represents a blog post
type Post struct {
	Base
	Title         string     `json:"title" gorm:"not null"`
	Content       string     `json:"content" gorm:"type:text;not null"`
	Excerpt       string     `json:"excerpt" gorm:"type:text"`
	FeaturedImage string     `json:"featured_image"`
	AuthorID      string     `json:"author_id" gorm:"not null"`
	AuthorName    string     `json:"author_name" gorm:"not null"`
	Categories    []string   `json:"categories" gorm:"type:text[]"`
	Tags          []string   `json:"tags" gorm:"type:text[]"`
	Published     bool       `json:"published" gorm:"not null;default:false"`
	PublishedAt   *time.Time `json:"published_at"`
}

// PostInput is used for creating and updating posts in the API layer
type PostInput struct {
	Title         string   `json:"title" binding:"required"`
	Content       string   `json:"content" binding:"required"`
	Excerpt       string   `json:"excerpt"`
	FeaturedImage string   `json:"featured_image"`
	Categories    []string `json:"categories"`
	Tags          []string `json:"tags"`
	Published     bool     `json:"published"`
}
