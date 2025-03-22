# Development Environment Setup Guide

This guide will help you set up your development environment for the BlogCommerce project, a combined blog and e-commerce website built with Hugo, PostgreSQL, and TinaCMS.

## Prerequisites

Make sure you have the following installed:

1. **Git**: [Download Git](https://git-scm.com/downloads)
2. **Go** (1.19+): [Download Go](https://golang.org/dl/)
3. **Node.js** (16+) and **npm**: [Download Node.js](https://nodejs.org/)
4. **PostgreSQL** (14+): [Download PostgreSQL](https://www.postgresql.org/download/)
5. **Hugo Extended** (0.110.0+): [Download Hugo](https://gohugo.io/installation/)

## Project Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/blogcommerce.git
cd blogcommerce
```

### 2. Set Up PostgreSQL Database

Create a new database and user for the project:

```bash
# Start PostgreSQL client
psql -U postgres

# In PostgreSQL prompt:
CREATE DATABASE blogcommerce;
CREATE USER blogcommerce_user WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE blogcommerce TO blogcommerce_user;
\q
```

### 3. Set Up Environment Variables

Create a `.env` file in the root directory:

```bash
# Create .env file
touch .env
```

Add the following environment variables to the `.env` file:

```
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=blogcommerce_user
DB_PASSWORD=your_password
DB_NAME=blogcommerce
DB_SSLMODE=disable

# JWT configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION_HOURS=24

# Server configuration
PORT=8080
GIN_MODE=debug

# Email configuration (optional for development)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM_NAME=BlogCommerce
EMAIL_FROM_ADDRESS=your-email@gmail.com
EMAIL_DEBUG=true

# TinaCMS configuration
TINA_TOKEN=your-tina-token
TINA_CLIENT_ID=your-tina-client-id

# API Base URL
API_BASE_URL=http://localhost:8080/api
```

Replace the placeholder values with your actual configuration.

### 4. API Setup

Install Go dependencies and set up the API:

```bash
# Navigate to the API directory
cd api

# Initialize Go modules (if not already initialized)
go mod tidy

# Run database migrations
go run cmd/migrations/main.go

# Start the API server
go run main.go
```

The API server should now be running at http://localhost:8080.

### 5. Hugo Setup

Set up the Hugo site:

```bash
# Return to the project root
cd ..

# Install Hugo theme dependencies
cd themes/blogcommerce
npm install

# Start Hugo development server
cd ../..
hugo server -D
```

The Hugo development server should now be running at http://localhost:1313.

### 6. TinaCMS Setup

Set up TinaCMS for local development:

```bash
# Install TinaCMS CLI
npm install -g @tinacms/cli

# Start TinaCMS local server
tinacms dev -c "hugo server -D"
```

TinaCMS should now be accessible through the Hugo site at http://localhost:1313/admin.

## Development Workflow

### API Development

When working on the API:

1. Make changes to Go files in the `api` directory
2. Test your changes with `go test ./...`
3. Run the API server with `go run main.go`

### Hugo Development

When working on the Hugo site:

1. Make changes to templates, styles, and content
2. View changes at http://localhost:1313
3. Build the site with `hugo`

### TinaCMS Development

When working with TinaCMS:

1. Configure content models in `tina/config.ts`
2. Create and edit content through the TinaCMS admin interface
3. Sync content with the database using the sync utility

## Directory Structure

```
blogcommerce/
├── api/                 # Go API code
│   ├── config/          # Configuration
│   ├── controllers/     # API controllers
│   ├── middleware/      # API middleware
│   ├── models/          # Data models
│   ├── routes/          # API routes
│   ├── utils/           # Utility functions
│   └── main.go          # Main entry point
├── content/             # Hugo content
│   ├── blog/            # Blog posts
│   ├── shop/            # Shop products
│   └── _index.md        # Home page
├── i18n/                # Internationalization files
├── layouts/             # Hugo templates
├── static/              # Static files
├── themes/              # Hugo themes
│   └── blogcommerce/    # Custom theme
│       ├── assets/      # Theme assets (SCSS, JS)
│       ├── layouts/     # Theme templates
│       └── static/      # Theme static files
├── tina/                # TinaCMS configuration
├── .env                 # Environment variables
├── .gitignore           # Git ignore file
├── config.yaml          # Hugo configuration
├── netlify.toml         # Netlify configuration
└── README.md            # Project documentation
```

## Testing

### API Testing

To run API tests:

```bash
cd api
go test ./... -v
```

For specific tests:

```bash
go test ./controllers -v
```

### Frontend Testing

For frontend component testing:

```bash
cd themes/blogcommerce
npm test
```

## Deployment

### Database Deployment

1. Create a PostgreSQL database on your hosting provider
2. Run the database migrations:

```bash
cd api
DATABASE_URL=your_production_database_url go run cmd/migrations/main.go
```

### API Deployment

The API can be deployed as a standalone Go application:

1. Build the API binary:

```bash
cd api
go build -o blogcommerce-api
```

2. Set up environment variables on your server
3. Run the binary:

```bash
./blogcommerce-api
```

Alternatively, you can deploy using Docker:

```bash
docker build -t blogcommerce-api ./api
docker run -p 8080:8080 --env-file .env blogcommerce-api
```

### Hugo Site Deployment

The Hugo site can be deployed to Netlify, Vercel, or any static site hosting:

1. Build the Hugo site:

```bash
hugo --minify
```

2. Deploy the `public` directory to your hosting provider

For Netlify deployment, push to your Git repository and configure the build settings:

- Build command: `hugo --minify`
- Publish directory: `public`

## Common Issues and Solutions

### Database Connection Issues

If you encounter database connection issues:

1. Ensure PostgreSQL is running:
   ```bash
   sudo service postgresql status
   ```

2. Check database credentials in `.env`

3. Test connection:
   ```bash
   psql -U blogcommerce_user -h localhost -d blogcommerce
   ```

### Hugo Build Errors

If Hugo build fails:

1. Check Hugo version:
   ```bash
   hugo version
   ```

2. Ensure you're using Hugo Extended

3. Clear the cache:
   ```bash
   hugo --gc
   ```

### TinaCMS Issues

If TinaCMS doesn't work properly:

1. Check TinaCMS token and client ID in `.env`

2. Restart TinaCMS server:
   ```bash
   tinacms dev -c "hugo server -D" --reset
   ```

3. Clear browser cache and cookies

## Additional Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [TinaCMS Documentation](https://tina.io/docs/)
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## Contributing

1. Fork the repository
2. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes
4. Run tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.