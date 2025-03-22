package utils

import (
	//"bufio"
	//"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	//"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// ContentType represents the type of content being synced
type ContentType string

const (
	BlogPost        ContentType = "blog_post"
	BlogCategory    ContentType = "blog_category"
	BlogTag         ContentType = "blog_tag"
	Product         ContentType = "product"
	ProductCategory ContentType = "product_category"
	Page            ContentType = "page"
)

// SyncConfig holds configuration for content synchronization
type SyncConfig struct {
	DB           *sql.DB
	ContentDir   string
	HugoDataDir  string
	TinaDataDir  string
	MediaDir     string
	BaseURL      string
	StaticSiteID string
}

// FrontMatter represents the YAML front matter of a content file
type FrontMatter struct {
	Title       string    `yaml:"title" json:"title"`
	Date        time.Time `yaml:"date" json:"date"`
	LastMod     time.Time `yaml:"lastmod" json:"lastmod"`
	Draft       bool      `yaml:"draft" json:"draft"`
	Slug        string    `yaml:"slug" json:"slug"`
	Description string    `yaml:"description" json:"description"`
	Categories  []string  `yaml:"categories" json:"categories"`
	Tags        []string  `yaml:"tags" json:"tags"`
	Author      string    `yaml:"author" json:"author"`
	Image       string    `yaml:"image" json:"image"`
	// Product specific fields
	Price      float64 `yaml:"price,omitempty" json:"price,omitempty"`
	SalePrice  float64 `yaml:"salePrice,omitempty" json:"salePrice,omitempty"`
	SKU        string  `yaml:"sku,omitempty" json:"sku,omitempty"`
	Stock      int     `yaml:"stock,omitempty" json:"stock,omitempty"`
	Featured   bool    `yaml:"featured,omitempty" json:"featured,omitempty"`
	Visible    bool    `yaml:"visible,omitempty" json:"visible,omitempty"`
	// Additional meta
	ID        int       `yaml:"id" json:"id"`
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `yaml:"updatedAt" json:"updatedAt"`
}

// ContentFile represents a content file with front matter and body
type ContentFile struct {
	FrontMatter FrontMatter `json:"frontMatter"`
	Body        string      `json:"body"`
	Path        string      `json:"path"`
}

// DBToHugo syncs content from the database to Hugo content files
func DBToHugo(config SyncConfig, contentType ContentType) error {
	switch contentType {
	case BlogPost:
		return syncBlogPostsFromDB(config)
	case BlogCategory:
		return syncBlogCategoriesFromDB(config)
	case BlogTag:
		return syncBlogTagsFromDB(config)
	case Product:
		return syncProductsFromDB(config)
	case ProductCategory:
		return syncProductCategoriesFromDB(config)
	case Page:
		return syncPagesFromDB(config)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// HugoDB syncs content from Hugo content files to the database
func HugoDB(config SyncConfig, contentType ContentType) error {
	switch contentType {
	case BlogPost:
		return syncBlogPostsToDB(config)
	case BlogCategory:
		return syncBlogCategoriesToDB(config)
	case BlogTag:
		return syncBlogTagsToDB(config)
	case Product:
		return syncProductsToDB(config)
	case ProductCategory:
		return syncProductCategoriesToDB(config)
	case Page:
		return syncPagesToDB(config)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// TinaToHugo syncs content from TinaCMS to Hugo content files
func TinaToHugo(config SyncConfig) error {
	// Read TinaCMS content directory
	files, err := ioutil.ReadDir(config.TinaDataDir)
	if err != nil {
		return err
	}

	// Process each file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Read TinaCMS content file
		filePath := filepath.Join(config.TinaDataDir, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse JSON content
		var contentFile ContentFile
		if err := json.Unmarshal(data, &contentFile); err != nil {
			return err
		}

		// Determine content type and path
		contentPath := contentFile.Path
		if contentPath == "" {
			// Derive path from file name
			baseName := strings.TrimSuffix(file.Name(), ".json")
			if strings.HasPrefix(baseName, "blog_post_") {
				contentPath = filepath.Join(config.ContentDir, "blog", contentFile.FrontMatter.Slug+".md")
			} else if strings.HasPrefix(baseName, "product_") {
				contentPath = filepath.Join(config.ContentDir, "shop", contentFile.FrontMatter.Slug+".md")
			} else if strings.HasPrefix(baseName, "page_") {
				contentPath = filepath.Join(config.ContentDir, contentFile.FrontMatter.Slug+".md")
			}
		}

		if contentPath == "" {
			continue // Skip if we can't determine the path
		}

		// Ensure directory exists
		dir := filepath.Dir(contentPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// Convert front matter to YAML
		frontMatter, err := yaml.Marshal(contentFile.FrontMatter)
		if err != nil {
			return err
		}

		// Write Hugo content file
		content := fmt.Sprintf("---\n%s---\n\n%s", string(frontMatter), contentFile.Body)
		if err := ioutil.WriteFile(contentPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// HugoToTina syncs content from Hugo content files to TinaCMS
func HugoToTina(config SyncConfig) error {
	// Content directories to scan
	contentDirs := []struct {
		dir         string
		contentType string
	}{
		{filepath.Join(config.ContentDir, "blog"), "blog_post"},
		{filepath.Join(config.ContentDir, "shop"), "product"},
		{filepath.Join(config.ContentDir), "page"},
	}

	// Process each content directory
	for _, contentDir := range contentDirs {
		// Skip if directory doesn't exist
		if _, err := os.Stat(contentDir.dir); os.IsNotExist(err) {
			continue
		}

		// Walk through the directory
		err := filepath.Walk(contentDir.dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories and non-markdown files
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			// Read content file
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// Parse front matter and body
			fm, body, err := parseMarkdownContent(string(content))
			if err != nil {
				return err
			}

			// Create content file
			contentFile := ContentFile{
				FrontMatter: fm,
				Body:        body,
				Path:        path,
			}

			// Generate Tina file name
			var tinaFileName string
			switch contentDir.contentType {
			case "blog_post":
				tinaFileName = fmt.Sprintf("blog_post_%s.json", fm.Slug)
			case "product":
				tinaFileName = fmt.Sprintf("product_%s.json", fm.Slug)
			case "page":
				tinaFileName = fmt.Sprintf("page_%s.json", fm.Slug)
			default:
				tinaFileName = fmt.Sprintf("content_%s.json", fm.Slug)
			}

			// Convert to JSON
			jsonData, err := json.MarshalIndent(contentFile, "", "  ")
			if err != nil {
				return err
			}

			// Ensure TinaCMS directory exists
			if err := os.MkdirAll(config.TinaDataDir, 0755); err != nil {
				return err
			}

			// Write TinaCMS content file
			tinaFilePath := filepath.Join(config.TinaDataDir, tinaFileName)
			if err := ioutil.WriteFile(tinaFilePath, jsonData, 0644); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// parseMarkdownContent parses markdown content into front matter and body
func parseMarkdownContent(content string) (FrontMatter, string, error) {
	var frontMatter FrontMatter
	var body string

	// Check if content has front matter
	if !strings.HasPrefix(content, "---") {
		return frontMatter, content, nil
	}

	// Find the end of front matter
	endIndex := strings.Index(content[3:], "---")
	if endIndex == -1 {
		return frontMatter, content, fmt.Errorf("invalid front matter format")
	}
	endIndex += 3 // Adjust for the offset in the substring

	// Extract front matter and body
	fmContent := content[3:endIndex]
	body = strings.TrimSpace(content[endIndex+3:])

	// Parse front matter
	err := yaml.Unmarshal([]byte(fmContent), &frontMatter)
	if err != nil {
		return frontMatter, body, fmt.Errorf("error parsing front matter: %v", err)
	}

	return frontMatter, body, nil
}

// Helper functions for syncing specific content types
func syncBlogPostsFromDB(config SyncConfig) error {
	// Query blog posts from database
	rows, err := config.DB.Query(`
		SELECT 
			p.id, p.title, p.slug, p.content, p.excerpt, 
			p.featured_image, p.published, p.published_at, 
			p.created_at, p.updated_at, u.name as author
		FROM blog_posts p
		LEFT JOIN users u ON p.author_id = u.id
		ORDER BY p.published_at DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Process each blog post
	for rows.Next() {
		var (
			id, title, slug, content, excerpt, featuredImage, author string
			published                                                bool
			publishedAt, createdAt, updatedAt                        time.Time
		)

		if err := rows.Scan(
			&id, &title, &slug, &content, &excerpt,
			&featuredImage, &published, &publishedAt,
			&createdAt, &updatedAt, &author,
		); err != nil {
			return err
		}

		// Get categories and tags for this post
		categories, err := getCategoriesForPost(config.DB, id)
		if err != nil {
			return err
		}

		tags, err := getTagsForPost(config.DB, id)
		if err != nil {
			return err
		}

		// Create front matter
		frontMatter := FrontMatter{
			Title:       title,
			Date:        publishedAt,
			LastMod:     updatedAt,
			Draft:       !published,
			Slug:        slug,
			Description: excerpt,
			Categories:  categories,
			Tags:        tags,
			Author:      author,
			Image:       featuredImage,
			//ID:          id,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		// Convert front matter to YAML
		fm, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Create content file
		contentFile := fmt.Sprintf("---\n%s---\n\n%s", string(fm), content)

		// Ensure directory exists
		postDir := filepath.Join(config.ContentDir, "blog")
		if err := os.MkdirAll(postDir, 0755); err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(postDir, slug+".md")
		if err := ioutil.WriteFile(filePath, []byte(contentFile), 0644); err != nil {
			return err
		}
	}

	return nil
}

// getTagsForPost fetches all tags for a blog post
func getTagsForPost(db *sql.DB, postID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.name
		FROM blog_tags t
		JOIN blog_post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
		ORDER BY t.name
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// getCategoriesForPost fetches all categories for a blog post
func getCategoriesForPost(db *sql.DB, postID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT c.name
		FROM blog_categories c
		JOIN blog_post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = $1
		ORDER BY c.name
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Implementation for other sync functions
func syncBlogCategoriesFromDB(config SyncConfig) error {
	// Query blog categories from database
	rows, err := config.DB.Query(`
		SELECT id, name, slug, description, created_at, updated_at
		FROM blog_categories
		ORDER BY name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Ensure directory exists
	catDir := filepath.Join(config.HugoDataDir, "categories")
	if err := os.MkdirAll(catDir, 0755); err != nil {
		return err
	}

	// Process each category
	for rows.Next() {
		var (
			id, name, slug, description string
			createdAt, updatedAt        time.Time
		)

		if err := rows.Scan(&id, &name, &slug, &description, &createdAt, &updatedAt); err != nil {
			return err
		}

		// Create front matter
		frontMatter := map[string]interface{}{
			"title":       name,
			"slug":        slug,
			"description": description,
			"id":          id,
			"createdAt":   createdAt,
			"updatedAt":   updatedAt,
		}

		// Convert to YAML
		data, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(catDir, slug+".yaml")
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func syncBlogTagsFromDB(config SyncConfig) error {
	// Query blog tags from database
	rows, err := config.DB.Query(`
		SELECT id, name, slug, created_at, updated_at
		FROM blog_tags
		ORDER BY name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Ensure directory exists
	tagDir := filepath.Join(config.HugoDataDir, "tags")
	if err := os.MkdirAll(tagDir, 0755); err != nil {
		return err
	}

	// Process each tag
	for rows.Next() {
		var (
			id, name, slug string
			createdAt, updatedAt time.Time
		)

		if err := rows.Scan(&id, &name, &slug, &createdAt, &updatedAt); err != nil {
			return err
		}

		// Create front matter
		frontMatter := map[string]interface{}{
			"title":     name,
			"slug":      slug,
			"id":        id,
			"createdAt": createdAt,
			"updatedAt": updatedAt,
		}

		// Convert to YAML
		data, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(tagDir, slug+".yaml")
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func syncProductsFromDB(config SyncConfig) error {
	// Query products from database
	rows, err := config.DB.Query(`
		SELECT 
			id, name, slug, description, price, sale_price,
			stock, sku, featured, visible, created_at, updated_at
		FROM products
		ORDER BY name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Process each product
	for rows.Next() {
		var (
			id, name, slug, description, sku string
			price, salePrice                 float64
			stock                            int
			featured, visible                bool
			createdAt, updatedAt             time.Time
		)

		if err := rows.Scan(
			&id, &name, &slug, &description, &price, &salePrice,
			&stock, &sku, &featured, &visible, &createdAt, &updatedAt,
		); err != nil {
			return err
		}

		// Get product images
		images, err := getProductImages(config.DB, id)
		if err != nil {
			return err
		}

		// Get product categories
		categories, err := getProductCategories(config.DB, id)
		if err != nil {
			return err
		}

		// Create front matter
		frontMatter := FrontMatter{
			Title:       name,
			Slug:        slug,
			Description: description,
			Price:       price,
			SalePrice:   salePrice,
			Stock:       stock,
			SKU:         sku,
			Featured:    featured,
			Visible:     visible,
			Categories:  categories,
			Image:       images[0], // Set first image as main image
			//ID:          id,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		// Convert front matter to YAML
		fm, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Create content with product description
		content := fmt.Sprintf("---\n%s---\n\n%s", string(fm), description)

		// Ensure directory exists
		productDir := filepath.Join(config.ContentDir, "shop")
		if err := os.MkdirAll(productDir, 0755); err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(productDir, slug+".md")
		if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// getProductImages fetches all images for a product
func getProductImages(db *sql.DB, productID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT image_url
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return nil, err
		}
		images = append(images, imageURL)
	}

	// If no images, add a placeholder
	if len(images) == 0 {
		images = append(images, "/images/placeholder.jpg")
	}

	return images, nil
}

// getProductCategories fetches all categories for a product
func getProductCategories(db *sql.DB, productID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT c.name
		FROM product_categories c
		JOIN product_category_items ci ON c.id = ci.category_id
		WHERE ci.product_id = $1
		ORDER BY c.name
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func syncProductCategoriesFromDB(config SyncConfig) error {
	// Query product categories from database
	rows, err := config.DB.Query(`
		SELECT id, name, slug, description, image, created_at, updated_at
		FROM product_categories
		ORDER BY name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Ensure directory exists
	catDir := filepath.Join(config.HugoDataDir, "product_categories")
	if err := os.MkdirAll(catDir, 0755); err != nil {
		return err
	}

	// Process each category
	for rows.Next() {
		var (
			id, name, slug, description, image string
			createdAt, updatedAt               time.Time
		)

		if err := rows.Scan(&id, &name, &slug, &description, &image, &createdAt, &updatedAt); err != nil {
			return err
		}

		// Create front matter
		frontMatter := map[string]interface{}{
			"title":       name,
			"slug":        slug,
			"description": description,
			"image":       image,
			"id":          id,
			"createdAt":   createdAt,
			"updatedAt":   updatedAt,
		}

		// Convert to YAML
		data, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(catDir, slug+".yaml")
		if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func syncPagesFromDB(config SyncConfig) error {
	// Query pages from database
	rows, err := config.DB.Query(`
		SELECT id, title, slug, content, published, updated_at, created_at
		FROM pages
		ORDER BY title
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Process each page
	for rows.Next() {
		var (
			id, title, slug, content string
			published                bool
			updatedAt, createdAt     time.Time
		)

		if err := rows.Scan(&id, &title, &slug, &content, &published, &updatedAt, &createdAt); err != nil {
			return err
		}

		// Create front matter
		frontMatter := FrontMatter{
			Title:     title,
			Slug:      slug,
			Draft:     !published,
			LastMod:   updatedAt,
			//ID:        id,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}

		// Convert front matter to YAML
		fm, err := yaml.Marshal(frontMatter)
		if err != nil {
			return err
		}

		// Create content file
		contentFile := fmt.Sprintf("---\n%s---\n\n%s", string(fm), content)

		// Ensure directory exists
		if err := os.MkdirAll(config.ContentDir, 0755); err != nil {
			return err
		}

		// Write file
		filePath := filepath.Join(config.ContentDir, slug+".md")
		if err := ioutil.WriteFile(filePath, []byte(contentFile), 0644); err != nil {
			return err
		}
	}

	return nil
}

func syncBlogPostsToDB(config SyncConfig) error {
	// Get all markdown files in the blog directory
	blogDir := filepath.Join(config.ContentDir, "blog")
	if _, err := os.Stat(blogDir); os.IsNotExist(err) {
		return nil // Skip if directory doesn't exist
	}

	files, err := ioutil.ReadDir(blogDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each markdown file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Read file
		filePath := filepath.Join(blogDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse front matter and content
		frontMatter, body, err := parseMarkdownContent(string(content))
		if err != nil {
			return err
		}

		// Check if post already exists
		var postID int
		err = tx.QueryRow("SELECT id FROM blog_posts WHERE slug = $1", frontMatter.Slug).Scan(&postID)
		if err == nil {
			// Post exists, update it
			_, err = tx.Exec(`
				UPDATE blog_posts
				SET title = $1, content = $2, excerpt = $3, featured_image = $4,
					published = $5, updated_at = NOW()
				WHERE id = $6
			`, frontMatter.Title, body, frontMatter.Description, frontMatter.Image, !frontMatter.Draft, postID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Post doesn't exist, create it
			// First, get author ID
			var authorID int
			err = tx.QueryRow("SELECT id FROM users WHERE name = $1", frontMatter.Author).Scan(&authorID)
			if err != nil {
				// Use the first admin as default author
				err = tx.QueryRow("SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&authorID)
				if err != nil {
					return err
				}
			}

			// Insert post
			err = tx.QueryRow(`
				INSERT INTO blog_posts (title, slug, content, excerpt, featured_image,
					published, author_id, published_at, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
				RETURNING id
			`, frontMatter.Title, frontMatter.Slug, body, frontMatter.Description,
				frontMatter.Image, !frontMatter.Draft, authorID, frontMatter.Date).Scan(&postID)
			if err != nil {
				return err
			}
		} else {
			return err
		}

		// Handle categories
		if len(frontMatter.Categories) > 0 {
			// First, clear existing categories
			_, err = tx.Exec("DELETE FROM blog_post_categories WHERE post_id = $1", postID)
			if err != nil {
				return err
			}

			// Add each category
			for _, categoryName := range frontMatter.Categories {
				// Find or create category
				var categoryID int
				err = tx.QueryRow("SELECT id FROM blog_categories WHERE name = $1", categoryName).Scan(&categoryID)
				if err == sql.ErrNoRows {
					// Create category with slugified name
					slug := slugify(categoryName)
					err = tx.QueryRow(`
						INSERT INTO blog_categories (name, slug, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())
						RETURNING id
					`, categoryName, slug).Scan(&categoryID)
					if err != nil {
						return err
					}
				} else if err != nil {
					return err
				}

				// Link post to category
				_, err = tx.Exec("INSERT INTO blog_post_categories (post_id, category_id) VALUES ($1, $2)",
					postID, categoryID)
				if err != nil {
					return err
				}
			}
		}

		// Handle tags
		if len(frontMatter.Tags) > 0 {
			// First, clear existing tags
			_, err = tx.Exec("DELETE FROM blog_post_tags WHERE post_id = $1", postID)
			if err != nil {
				return err
			}

			// Add each tag
			for _, tagName := range frontMatter.Tags {
				// Find or create tag
				var tagID int
				err = tx.QueryRow("SELECT id FROM blog_tags WHERE name = $1", tagName).Scan(&tagID)
				if err == sql.ErrNoRows {
					// Create tag with slugified name
					slug := slugify(tagName)
					err = tx.QueryRow(`
						INSERT INTO blog_tags (name, slug, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())
						RETURNING id
					`, tagName, slug).Scan(&tagID)
					if err != nil {
						return err
					}
				} else if err != nil {
					return err
				}

				// Link post to tag
				_, err = tx.Exec("INSERT INTO blog_post_tags (post_id, tag_id) VALUES ($1, $2)",
					postID, tagID)
				if err != nil {
					return err
				}
			}
		}
	}

	// Commit transaction
	return tx.Commit()
}

func syncBlogCategoriesToDB(config SyncConfig) error {
	// Get all YAML files in the categories directory
	catDir := filepath.Join(config.HugoDataDir, "categories")
	if _, err := os.Stat(catDir); os.IsNotExist(err) {
		return nil // Skip if directory doesn't exist
	}

	files, err := ioutil.ReadDir(catDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each YAML file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Read file
		filePath := filepath.Join(catDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse YAML
		var category map[string]interface{}
		if err := yaml.Unmarshal(content, &category); err != nil {
			return err
		}

		// Extract fields
		title := getStringValue(category, "title")
		slug := getStringValue(category, "slug")
		description := getStringValue(category, "description")
		id := getStringValue(category, "id")

		// Check if category already exists
		var categoryID int
		if id != "" {
			err = tx.QueryRow("SELECT id FROM blog_categories WHERE id = $1", id).Scan(&categoryID)
		} else {
			err = tx.QueryRow("SELECT id FROM blog_categories WHERE slug = $1", slug).Scan(&categoryID)
		}

		if err == nil {
			// Category exists, update it
			_, err = tx.Exec(`
				UPDATE blog_categories
				SET name = $1, description = $2, updated_at = NOW()
				WHERE id = $3
			`, title, description, categoryID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Category doesn't exist, create it
			_, err = tx.Exec(`
				INSERT INTO blog_categories (name, slug, description, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
			`, title, slug, description)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

func syncBlogTagsToDB(config SyncConfig) error {
	// Get all YAML files in the tags directory
	tagDir := filepath.Join(config.HugoDataDir, "tags")
	if _, err := os.Stat(tagDir); os.IsNotExist(err) {
		return nil // Skip if directory doesn't exist
	}

	files, err := ioutil.ReadDir(tagDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each YAML file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Read file
		filePath := filepath.Join(tagDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse YAML
		var tag map[string]interface{}
		if err := yaml.Unmarshal(content, &tag); err != nil {
			return err
		}

		// Extract fields
		title := getStringValue(tag, "title")
		slug := getStringValue(tag, "slug")
		id := getStringValue(tag, "id")

		// Check if tag already exists
		var tagID int
		if id != "" {
			err = tx.QueryRow("SELECT id FROM blog_tags WHERE id = $1", id).Scan(&tagID)
		} else {
			err = tx.QueryRow("SELECT id FROM blog_tags WHERE slug = $1", slug).Scan(&tagID)
		}

		if err == nil {
			// Tag exists, update it
			_, err = tx.Exec(`
				UPDATE blog_tags
				SET name = $1, updated_at = NOW()
				WHERE id = $2
			`, title, tagID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Tag doesn't exist, create it
			_, err = tx.Exec(`
				INSERT INTO blog_tags (name, slug, created_at, updated_at)
				VALUES ($1, $2, NOW(), NOW())
			`, title, slug)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

func syncProductsToDB(config SyncConfig) error {
	// Get all markdown files in the shop directory
	shopDir := filepath.Join(config.ContentDir, "shop")
	if _, err := os.Stat(shopDir); os.IsNotExist(err) {
		return nil // Skip if directory doesn't exist
	}

	files, err := ioutil.ReadDir(shopDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each markdown file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Read file
		filePath := filepath.Join(shopDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse front matter and content
		frontMatter, _, err := parseMarkdownContent(string(content))
		if err != nil {
			return err
		}

		// Check if product already exists
		var productID int
		err = tx.QueryRow("SELECT id FROM products WHERE slug = $1", frontMatter.Slug).Scan(&productID)
		if err == nil {
			// Product exists, update it
			_, err = tx.Exec(`
				UPDATE products
				SET name = $1, description = $2, price = $3, sale_price = $4,
					stock = $5, sku = $6, featured = $7, visible = $8, updated_at = NOW()
				WHERE id = $9
			`, frontMatter.Title, frontMatter.Description, frontMatter.Price, frontMatter.SalePrice,
				frontMatter.Stock, frontMatter.SKU, frontMatter.Featured, frontMatter.Visible, productID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Product doesn't exist, create it
			err = tx.QueryRow(`
				INSERT INTO products (name, slug, description, price, sale_price,
					stock, sku, featured, visible, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
				RETURNING id
			`, frontMatter.Title, frontMatter.Slug, frontMatter.Description, frontMatter.Price,
				frontMatter.SalePrice, frontMatter.Stock, frontMatter.SKU, frontMatter.Featured, 
				frontMatter.Visible).Scan(&productID)
			if err != nil {
				return err
			}
		} else {
			return err
		}

		// Handle product image
		if frontMatter.Image != "" {
			// Check if product already has a primary image
			var imageID int
			err = tx.QueryRow(`
				SELECT id FROM product_images 
				WHERE product_id = $1 AND is_primary = true
				LIMIT 1
			`, productID).Scan(&imageID)

			if err == nil {
				// Update existing primary image
				_, err = tx.Exec(`
					UPDATE product_images
					SET image_url = $1, updated_at = NOW()
					WHERE id = $2
				`, frontMatter.Image, imageID)
				if err != nil {
					return err
				}
			} else if err == sql.ErrNoRows {
				// No primary image, create one
				_, err = tx.Exec(`
					INSERT INTO product_images (product_id, image_url, alt_text, is_primary, sort_order, created_at)
					VALUES ($1, $2, $3, true, 0, NOW())
				`, productID, frontMatter.Image, frontMatter.Title)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Handle product categories
		if len(frontMatter.Categories) > 0 {
			// First, clear existing categories
			_, err = tx.Exec("DELETE FROM product_category_items WHERE product_id = $1", productID)
			if err != nil {
				return err
			}

			// Add each category
			for _, categoryName := range frontMatter.Categories {
				// Find or create category
				var categoryID int
				err = tx.QueryRow("SELECT id FROM product_categories WHERE name = $1", categoryName).Scan(&categoryID)
				if err == sql.ErrNoRows {
					// Create category with slugified name
					slug := slugify(categoryName)
					err = tx.QueryRow(`
						INSERT INTO product_categories (name, slug, created_at, updated_at)
						VALUES ($1, $2, NOW(), NOW())
						RETURNING id
					`, categoryName, slug).Scan(&categoryID)
					if err != nil {
						return err
					}
				} else if err != nil {
					return err
				}

				// Link product to category
				_, err = tx.Exec("INSERT INTO product_category_items (product_id, category_id) VALUES ($1, $2)",
					productID, categoryID)
				if err != nil {
					return err
				}
			}
		}
	}

	// Commit transaction
	return tx.Commit()
}

func syncProductCategoriesToDB(config SyncConfig) error {
	// Get all YAML files in the product_categories directory
	catDir := filepath.Join(config.HugoDataDir, "product_categories")
	if _, err := os.Stat(catDir); os.IsNotExist(err) {
		return nil // Skip if directory doesn't exist
	}

	files, err := ioutil.ReadDir(catDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each YAML file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Read file
		filePath := filepath.Join(catDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse YAML
		var category map[string]interface{}
		if err := yaml.Unmarshal(content, &category); err != nil {
			return err
		}

		// Extract fields
		title := getStringValue(category, "title")
		slug := getStringValue(category, "slug")
		description := getStringValue(category, "description")
		image := getStringValue(category, "image")
		id := getStringValue(category, "id")

		// Check if category already exists
		var categoryID int
		if id != "" {
			err = tx.QueryRow("SELECT id FROM product_categories WHERE id = $1", id).Scan(&categoryID)
		} else {
			err = tx.QueryRow("SELECT id FROM product_categories WHERE slug = $1", slug).Scan(&categoryID)
		}

		if err == nil {
			// Category exists, update it
			_, err = tx.Exec(`
				UPDATE product_categories
				SET name = $1, description = $2, image = $3, updated_at = NOW()
				WHERE id = $4
			`, title, description, image, categoryID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Category doesn't exist, create it
			_, err = tx.Exec(`
				INSERT INTO product_categories (name, slug, description, image, created_at, updated_at)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, title, slug, description, image)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

func syncPagesToDB(config SyncConfig) error {
	// Get all markdown files in the content directory (excluding blog and shop)
	contentDir := config.ContentDir
	files, err := ioutil.ReadDir(contentDir)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process each markdown file
	for _, file := range files {
		// Skip directories and non-markdown files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		
		// Skip special files like _index.md
		if strings.HasPrefix(file.Name(), "_") {
			continue
		}

		// Read file
		filePath := filepath.Join(contentDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Parse front matter and content
		frontMatter, body, err := parseMarkdownContent(string(content))
		if err != nil {
			return err
		}

		// Check if page already exists
		var pageID int
		err = tx.QueryRow("SELECT id FROM pages WHERE slug = $1", frontMatter.Slug).Scan(&pageID)
		if err == nil {
			// Page exists, update it
			_, err = tx.Exec(`
				UPDATE pages
				SET title = $1, content = $2, published = $3, updated_at = NOW()
				WHERE id = $4
			`, frontMatter.Title, body, !frontMatter.Draft, pageID)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Page doesn't exist, create it
			_, err = tx.Exec(`
				INSERT INTO pages (title, slug, content, published, created_at, updated_at)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, frontMatter.Title, frontMatter.Slug, body, !frontMatter.Draft)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

// Helper functions

// getStringValue safely extracts a string value from a map
func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// slugify creates a URL-friendly slug from a string
func slugify(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)
	
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	
	// Trim hyphens from beginning and end
	slug = strings.Trim(slug, "-")
	
	return slug
} 