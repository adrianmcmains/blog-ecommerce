package repository

import (
    "context"
    
    "github.com/adrianmcmains/blog-ecommerce/internal/models"
	"gorm.io/gorm"
) 

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id string) (*models.Post, error)
	GetAll(ctx context.Context, page, limit int) ([]*models.Post, error)
	GetByCategory(ctx context.Context, category string, page, limit int) ([]*models.Post, error)
	GetByAuthor(ctx context.Context, authorID string, page, limit int) ([]*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

// UserRepository defines the methods for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type Users interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, id string, user *models.User) error
}

type Posts interface {
    Create(ctx context.Context, post *models.Post) error
    GetByID(ctx context.Context, id string) (*models.Post, error)
    List(ctx context.Context, filter models.PostFilter) ([]models.Post, int64, error)
    Update(ctx context.Context, id string, post *models.Post) error
    Delete(ctx context.Context, id string) error
}

type Products interface {
    Create(ctx context.Context, product *models.Product) error
    GetByID(ctx context.Context, id string) (*models.Product, error)
    List(ctx context.Context, filter models.ProductFilter) ([]models.Product, int64, error)
    Update(ctx context.Context, id string, product *models.Product) error
    Delete(ctx context.Context, id string) error
    UpdateStock(ctx context.Context, id string, quantity int) error
}

type CartItems interface {
    Create(ctx context.Context, item *models.CartItem) error
    GetByID(ctx context.Context, id string) (*models.CartItem, error)
    GetUserCart(ctx context.Context, userID string) ([]models.CartItem, error)
    Update(ctx context.Context, id string, item *models.CartItem) error
    Delete(ctx context.Context, id string) error
    DeleteUserCart(ctx context.Context, userID string) error
}

type Orders interface {
    Create(ctx context.Context, order *models.Order) error
    GetByID(ctx context.Context, id string) (*models.Order, error)
    List(ctx context.Context, filter models.OrderFilter) ([]models.Order, int64, error)
    Update(ctx context.Context, id string, order *models.Order) error
}

type Repository struct {
    Users     Users
    Posts     Posts
    Products  Products
    CartItems CartItems
    Orders    Orders
}

// TokenRepository handles token data storage operations
type TokenRepository interface {
	Create(ctx context.Context, token *models.Token) error
	GetByID(ctx context.Context, id string) (*models.Token, error)
	GetByAccessToken(ctx context.Context, accessToken string) (*models.Token, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Token, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Token, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

// PostsRepo implements the Posts interface
type PostsRepo struct {
    db *gorm.DB
}

// NewPostsRepo creates a new PostsRepo
func NewPostsRepo(db *gorm.DB) Posts {
    return &PostsRepo{
        db: db,
    }
}

// Create implements the Create method of the Posts interface
func (r *PostsRepo) Create(ctx context.Context, post *models.Post) error {
    return r.db.WithContext(ctx).Create(post).Error
}

// GetByID implements the GetByID method of the Posts interface
func (r *PostsRepo) GetByID(ctx context.Context, id string) (*models.Post, error) {
    var post models.Post
    err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &post, nil
}

// List implements the List method of the Posts interface
func (r *PostsRepo) List(ctx context.Context, filter models.PostFilter) ([]models.Post, int64, error) {
    var posts []models.Post
    var count int64

    // Start with a base query
    query := r.db.WithContext(ctx)

    // Apply filters if needed
    // Example: if filter.AuthorID is set
    // if filter.AuthorID != "" {
    //     query = query.Where("author_id = ?", filter.AuthorID)
    // }

    // Count total records
    if err := query.Model(&models.Post{}).Count(&count).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination if needed
    // if filter.Limit > 0 {
    //     query = query.Limit(filter.Limit).Offset(filter.Offset)
    // }

    // Fetch posts
    err := query.Find(&posts).Error
    if err != nil {
        return nil, 0, err
    }

    return posts, count, nil
}

// Update implements the Update method of the Posts interface
func (r *PostsRepo) Update(ctx context.Context, id string, post *models.Post) error {
    return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", id).Updates(post).Error
}

// Delete implements the Delete method of the Posts interface
func (r *PostsRepo) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&models.Post{}, "id = ?", id).Error
}

func NewRepository(db *gorm.DB) *Repository {
    return &Repository{
        Users:     NewUsersRepo(db),
        Posts:     NewPostsRepo(db),
        Products:  NewProductsRepo(db),
        CartItems: NewCartItemsRepo(db),
        Orders:    NewOrdersRepo(db),
    }
}

// tokenRepositoryImpl implements TokenRepository
type tokenRepositoryImpl struct {
    db *gorm.DB
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(db *gorm.DB) TokenRepository {
    return &tokenRepositoryImpl{db: db}
}

// Create implements the Create method of TokenRepository
func (r *tokenRepositoryImpl) Create(ctx context.Context, token *models.Token) error {
    return r.db.WithContext(ctx).Create(token).Error
}

// GetByID implements the GetByID method of TokenRepository
func (r *tokenRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Token, error) {
    var token models.Token
    err := r.db.WithContext(ctx).First(&token, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

// GetByAccessToken implements the GetByAccessToken method of TokenRepository
func (r *tokenRepositoryImpl) GetByAccessToken(ctx context.Context, accessToken string) (*models.Token, error) {
    var token models.Token
    err := r.db.WithContext(ctx).First(&token, "access_token = ?", accessToken).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

// GetByRefreshToken implements the GetByRefreshToken method of TokenRepository
func (r *tokenRepositoryImpl) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Token, error) {
    var token models.Token
    err := r.db.WithContext(ctx).First(&token, "refresh_token = ?", refreshToken).Error
    if err != nil {
        return nil, err
    }
    return &token, nil
}

// GetByUserID implements the GetByUserID method of TokenRepository
func (r *tokenRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*models.Token, error) {
    var tokens []*models.Token
    err := r.db.WithContext(ctx).Find(&tokens, "user_id = ?", userID).Error
    if err != nil {
        return nil, err
    }
    return tokens, nil
}

// Revoke implements the Revoke method of TokenRepository
func (r *tokenRepositoryImpl) Revoke(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Model(&models.Token{}).Where("id = ?", id).Update("revoked", true).Error
}

// RevokeAllForUser implements the RevokeAllForUser method of TokenRepository
func (r *tokenRepositoryImpl) RevokeAllForUser(ctx context.Context, userID string) error {
    return r.db.WithContext(ctx).Model(&models.Token{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

// DeleteExpired implements the DeleteExpired method of TokenRepository
func (r *tokenRepositoryImpl) DeleteExpired(ctx context.Context) error {
    return r.db.WithContext(ctx).Where("expires_at < ?", ctx.Value("now")).Delete(&models.Token{}).Error
}