package util

import (
    "crypto/rand"
    "encoding/base64"
    //"regexp"
    "strings"
)

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(s string) string {
    // Convert to lowercase and replace spaces with hyphens
    slug := strings.ToLower(s)
    slug = strings.ReplaceAll(slug, " ", "-")
    
    // Remove special characters
    slug = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
            return r
        }
        return -1
    }, slug)

    return slug
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, length int) string {
    if len(s) <= length {
        return s
    }
    return s[:length] + "..."
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// RemoveDuplicates removes duplicates from a string slice
func RemoveDuplicates(slice []string) []string {
    keys := make(map[string]bool)
    result := []string{}
    
    for _, entry := range slice {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            result = append(result, entry)
        }
    }
    
    return result
}