package service

import (
    "time"
    "context"
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
    "github.com/adrianmcmains/blog-ecommerce/internal/repository"
)

// Service struct that combines all services
type Service struct {
    Auth *AuthService
    Blog *BlogService
    Shop *ShopService
}

// NewService creates a new Service with all required dependencies
func NewService(
    repos *repository.Repository,
    tokenRepo repository.TokenRepository,
    jwtSecret string,
    jwtTTL time.Duration,
    refreshTTL time.Duration,
) *Service {
    // Create UserRepository adapter
    userRepo := &UserRepoAdapter{repos.Users}
    
    // Create PostRepository adapter
    postRepo := &PostRepoAdapter{repos.Posts}
    
    return &Service{
        Auth: NewAuthService(userRepo, tokenRepo, jwtSecret, jwtTTL, refreshTTL),
        Blog: NewBlogService(postRepo, userRepo),
        Shop: NewShopService(repos.Products, repos.CartItems, repos.Orders),
    }
}

// UserRepoAdapter adapts repository.Users to repository.UserRepository
type UserRepoAdapter struct {
    Users repository.Users
}

func (a *UserRepoAdapter) Create(ctx context.Context, user *models.User) error {
    return a.Users.Create(ctx, user)
}

func (a *UserRepoAdapter) GetByID(ctx context.Context, id string) (*models.User, error) {
    return a.Users.GetByID(ctx, id)
}

func (a *UserRepoAdapter) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    return a.Users.GetByEmail(ctx, email)
}

func (a *UserRepoAdapter) Update(ctx context.Context, user *models.User) error {
    return a.Users.Update(ctx, user.ID.String(), user)
}

// PostRepoAdapter adapts repository.Posts to repository.PostRepository
type PostRepoAdapter struct {
    Posts repository.Posts
}

func (a *PostRepoAdapter) Create(ctx context.Context, post *models.Post) error {
    return a.Posts.Create(ctx, post)
}

func (a *PostRepoAdapter) GetByID(ctx context.Context, id string) (*models.Post, error) {
    return a.Posts.GetByID(ctx, id)
}

func (a *PostRepoAdapter) GetAll(ctx context.Context, page, limit int) ([]*models.Post, error) {
    // Create a filter without relying on specific field names
    // Using a dynamic approach since we don't know the exact field names
    filter := models.PostFilter{}
    
    // Pass pagination via the List method
    posts, _, err := a.Posts.List(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // Handle pagination in memory if we have to
    startIdx := (page - 1) * limit
    endIdx := startIdx + limit
    if startIdx >= len(posts) {
        return []*models.Post{}, nil
    }
    if endIdx > len(posts) {
        endIdx = len(posts)
    }
    
    // Apply pagination
    paginatedPosts := posts[startIdx:endIdx]
    
    // Convert to the required return type
    result := make([]*models.Post, len(paginatedPosts))
    for i := range paginatedPosts {
        post := paginatedPosts[i]
        result[i] = &post
    }
    return result, nil
}

func (a *PostRepoAdapter) GetByCategory(ctx context.Context, category string, page, limit int) ([]*models.Post, error) {
    // Create a basic filter - we'll assume it has a Category field
    filter := models.PostFilter{}
    
    // Use reflection to set the Category field regardless of its exact name
    // Here we're assuming PostFilter.Category exists
    // Alternatively, you could modify the Posts.List implementation to accept a category parameter
    
    // Apply filter
    posts, _, err := a.Posts.List(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // Filter by category in memory
    var categoryPosts []models.Post
    for _, post := range posts {
        for _, postCategory := range post.Categories {
            if postCategory == category {
                categoryPosts = append(categoryPosts, post)
                break
            }
        }
    }
    
    // Handle pagination
    startIdx := (page - 1) * limit
    endIdx := startIdx + limit
    if startIdx >= len(categoryPosts) {
        return []*models.Post{}, nil
    }
    if endIdx > len(categoryPosts) {
        endIdx = len(categoryPosts)
    }
    
    // Apply pagination
    paginatedPosts := categoryPosts[startIdx:endIdx]
    
    // Convert to the required return type
    result := make([]*models.Post, len(paginatedPosts))
    for i := range paginatedPosts {
        post := paginatedPosts[i]
        result[i] = &post
    }
    return result, nil
}

func (a *PostRepoAdapter) GetByAuthor(ctx context.Context, authorID string, page, limit int) ([]*models.Post, error) {
    // Create a basic filter
    filter := models.PostFilter{}
    
    // Get all posts and filter by author in memory
    posts, _, err := a.Posts.List(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // Filter by author
    var authorPosts []models.Post
    for _, post := range posts {
        if post.AuthorID == authorID {
            authorPosts = append(authorPosts, post)
        }
    }
    
    // Handle pagination
    startIdx := (page - 1) * limit
    endIdx := startIdx + limit
    if startIdx >= len(authorPosts) {
        return []*models.Post{}, nil
    }
    if endIdx > len(authorPosts) {
        endIdx = len(authorPosts)
    }
    
    // Apply pagination
    paginatedPosts := authorPosts[startIdx:endIdx]
    
    // Convert to the required return type
    result := make([]*models.Post, len(paginatedPosts))
    for i := range paginatedPosts {
        post := paginatedPosts[i]
        result[i] = &post
    }
    return result, nil
}

func (a *PostRepoAdapter) Update(ctx context.Context, post *models.Post) error {
    return a.Posts.Update(ctx, post.ID.String(), post)
}

func (a *PostRepoAdapter) Delete(ctx context.Context, id string) error {
    return a.Posts.Delete(ctx, id)
}

func (a *PostRepoAdapter) Count(ctx context.Context) (int64, error) {
    _, count, err := a.Posts.List(ctx, models.PostFilter{})
    return count, err
}