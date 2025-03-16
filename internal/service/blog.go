package service

import (
	"context"
	"errors"
	"time"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/repository"
)

// PostInput represents the input for creating or updating a post
type PostInput struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Excerpt       string   `json:"excerpt"`
	FeaturedImage string   `json:"featured_image"`
	Categories    []string `json:"categories"`
	Tags          []string `json:"tags"`
	Published     bool     `json:"published"`
}

// Common errors
var (
	ErrPostNotFound = errors.New("post not found")
	ErrNotAuthorized = errors.New("not authorized to perform this action")
)

// BlogService handles blog-related business logic
type BlogService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

// NewBlogService creates a new blog service
func NewBlogService(postRepo repository.PostRepository, userRepo repository.UserRepository) *BlogService {
	return &BlogService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new blog post
func (s *BlogService) CreatePost(ctx context.Context, input PostInput, authorID string) (*models.Post, error) {
	// Get author details
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	
	authorName := author.FirstName + " " + author.LastName

	// Create new post
	post := &models.Post{
		Base: models.Base{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:         input.Title,
		Content:       input.Content,
		Excerpt:       input.Excerpt,
		FeaturedImage: input.FeaturedImage,
		AuthorID:      authorID,
		AuthorName:    authorName,
		Categories:    input.Categories,
		Tags:          input.Tags,
		Published:     input.Published,
	}

	// Handle published status
	if input.Published {
		now := time.Now()
		post.PublishedAt = &now
	}

	// Save post to repository
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// GetPost retrieves a post by ID
func (s *BlogService) GetPost(ctx context.Context, id string) (*models.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}

// ListPosts retrieves all posts with filtering and pagination
func (s *BlogService) ListPosts(ctx context.Context, page, limit int, category, authorID string) ([]*models.Post, int64, error) {
	var posts []*models.Post
	var err error

	// Apply filters if provided
	if category != "" {
		posts, err = s.postRepo.GetByCategory(ctx, category, page, limit)
	} else if authorID != "" {
		posts, err = s.postRepo.GetByAuthor(ctx, authorID, page, limit)
	} else {
		posts, err = s.postRepo.GetAll(ctx, page, limit)
	}

	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	count, err := s.postRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// UpdatePost updates an existing post
func (s *BlogService) UpdatePost(ctx context.Context, id string, input PostInput, userID string) (*models.Post, error) {
	// Get the existing post
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// Check if user is the author of the post
	if post.AuthorID != userID {
		return nil, ErrNotAuthorized
	}

	// Update fields
	post.Title = input.Title
	post.Content = input.Content
	post.UpdatedAt = time.Now()

	// Update optional fields if provided
	if input.Excerpt != "" {
		post.Excerpt = input.Excerpt
	}
	if input.FeaturedImage != "" {
		post.FeaturedImage = input.FeaturedImage
	}
	if len(input.Categories) > 0 {
		post.Categories = input.Categories
	}
	if len(input.Tags) > 0 {
		post.Tags = input.Tags
	}

	// Handle publishing status change
	if post.Published != input.Published {
		post.Published = input.Published
		if input.Published {
			now := time.Now()
			post.PublishedAt = &now
		}
	}

	// Save updated post
	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost deletes a post
func (s *BlogService) DeletePost(ctx context.Context, id string, userID string) error {
	// Get the post to check authorization
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return ErrPostNotFound
	}

	// Check if user is the author of the post
	if post.AuthorID != userID {
		return ErrNotAuthorized
	}

	// Delete the post
	return s.postRepo.Delete(ctx, id)
}