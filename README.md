# BlogCommerce

A modern combined blog and e-commerce platform built with Hugo, PostgreSQL, and TinaCMS.

![BlogCommerce Demo](https://via.placeholder.com/1200x600?text=BlogCommerce)

## Features

### Blog Features
- Clean, modern design based on Hugo Reporter theme
- Category and tag organization
- Author profiles
- SEO optimization
- Comments system
- Related posts

### E-commerce Features
- Product catalog with categories
- Shopping cart with local storage persistence
- Secure checkout process
- Order management
- Payment gateway integration (Eversend, PayPal)

### CMS Features
- TinaCMS integration for content management
- User-friendly admin interface
- Media library management
- Markdown editor with preview

### Technical Features
- PostgreSQL database for dynamic content
- REST API built with Go
- JWT authentication
- Responsive design
- Internationalization support
- Accessibility compliance (WCAG)
- SEO optimization
- Email notifications
- Analytics integration

## Technology Stack

- **Frontend**: Hugo, JavaScript, SCSS
- **Backend**: Go, Gin Framework
- **Database**: PostgreSQL
- **CMS**: TinaCMS
- **Authentication**: JWT
- **Deployment**: Netlify, Docker

## Getting Started

For detailed setup instructions, refer to [DEV_SETUP.md](DEV_SETUP.md).

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/blogcommerce.git
   cd blogcommerce
   ```

2. Set up the environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Start the API:
   ```bash
   cd api
   go mod tidy
   go run main.go
   ```

4. Start Hugo:
   ```bash
   cd ..
   hugo server -D
   ```

5. Visit http://localhost:1313 to see the site

## Documentation

- [Development Setup](DEV_SETUP.md)
- [API Documentation](docs/API.md)
- [Database Schema](docs/SCHEMA.md)
- [Content Management Guide](docs/CMS.md)
- [Deployment Guide](docs/DEPLOYMENT.md)

## Project Structure

The project is organized into several main components:

```
blogcommerce/
├── api/            # Go API backend
├── content/        # Hugo content files
├── i18n/           # Internationalization files
├── layouts/        # Hugo templates
├── static/         # Static files
├── themes/         # Hugo themes
├── tina/           # TinaCMS configuration
```

## Screenshots

### Home Page
![Home Page](https://via.placeholder.com/800x400?text=Home+Page)

### Blog Page
![Blog Page](https://via.placeholder.com/800x400?text=Blog+Page)

### Shop Page
![Shop Page](https://via.placeholder.com/800x400?text=Shop+Page)

### Admin Dashboard
![Admin Dashboard](https://via.placeholder.com/800x400?text=Admin+Dashboard)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Hugo](https://gohugo.io/) for the static site generation
- [TinaCMS](https://tina.io/) for the content management system
- [Gin](https://gin-gonic.com/) for the Go web framework
- [PostgreSQL](https://www.postgresql.org/) for the database
- [Netlify](https://www.netlify.com/) for hosting and deployment