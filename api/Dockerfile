# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o blogcommerce-api .

# Final stage
FROM alpine:latest

# Add necessary packages
RUN apk --no-cache add ca-certificates postgresql-client mailcap

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/blogcommerce-api .

# Create directories for templates and static files
RUN mkdir -p /app/templates /app/static

# Copy email templates and other necessary files
COPY templates/ /app/templates/
COPY static/ /app/static/

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release \
    PORT=8080

# Command to run the executable
ENTRYPOINT ["./blogcommerce-api"]