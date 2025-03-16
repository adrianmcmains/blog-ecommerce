package repository

import (
	"context"
	"errors"
	"time"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoPostRepository implements PostRepository using MongoDB
type MongoPostRepository struct {
	collection *mongo.Collection
}

// NewMongoPostRepository creates a new MongoDB-based post repository
func NewMongoPostRepository(db *mongo.Database) *MongoPostRepository {
	collection := db.Collection("posts")
	
	// Create indexes
	titleIndex := mongo.IndexModel{
		Keys: bson.M{"title": 1},
	}
	
	slugIndex := mongo.IndexModel{
		Keys:    bson.M{"slug": 1},
		Options: options.Index().SetUnique(true),
	}
	
	authorIndex := mongo.IndexModel{
		Keys: bson.M{"author_id": 1},
	}
	
	categoryIndex := mongo.IndexModel{
		Keys: bson.M{"categories": 1},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		titleIndex,
		slugIndex,
		authorIndex,
		categoryIndex,
	})
	if err != nil {
		// Log the error but don't fail
		// log.Printf("Failed to create post indexes: %v", err)
	}
	
	return &MongoPostRepository{
		collection: collection,
	}
}

// Create inserts a new post into the database
func (r *MongoPostRepository) Create(ctx context.Context, post *models.Post) error {
	_, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("post with this slug already exists")
		}
		return err
	}
	return nil
}

// GetByID retrieves a post by ID
func (r *MongoPostRepository) GetByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

// GetAll retrieves all posts with pagination
func (r *MongoPostRepository) GetAll(ctx context.Context, page, limit int) ([]*models.Post, error) {
	skip := (page - 1) * limit
	
	options := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, bson.M{}, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var posts []*models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	
	return posts, nil
}

// Update updates a post in the database
func (r *MongoPostRepository) Update(ctx context.Context, post *models.Post) error {
	post.UpdatedAt = time.Now()
	
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": post.ID}, post)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("post not found")
	}
	return nil
}

// Delete removes a post from the database
func (r *MongoPostRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("post not found")
	}
	return nil
}

// GetByCategory retrieves posts by category
func (r *MongoPostRepository) GetByCategory(ctx context.Context, category string, page, limit int) ([]*models.Post, error) {
	skip := (page - 1) * limit
	
	filter := bson.M{"categories": category}
	
	options := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var posts []*models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	
	return posts, nil
}

// GetByAuthor retrieves posts by author
func (r *MongoPostRepository) GetByAuthor(ctx context.Context, authorID string, page, limit int) ([]*models.Post, error) {
	skip := (page - 1) * limit
	
	filter := bson.M{"author_id": authorID}
	
	options := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var posts []*models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}
	
	return posts, nil
}

// Count returns the total number of posts
func (r *MongoPostRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}