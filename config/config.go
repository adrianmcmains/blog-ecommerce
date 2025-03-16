package config

import (
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL string
    Port        string
    JWTSecret   string
    Environment string
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, err
    }

    config := &Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        Port:        os.Getenv("PORT"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
        Environment: os.Getenv("ENVIRONMENT"),
    }

    // Set defaults if not provided
    if config.Port == "" {
        config.Port = "8080"
    }

    if config.Environment == "" {
        config.Environment = "development"
    }

    return config, nil
}