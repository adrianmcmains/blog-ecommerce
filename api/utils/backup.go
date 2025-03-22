package utils

import (
	"bytes"
	"database/sql"
	//"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupConfig holds configuration for backup operations
type BackupConfig struct {
	DB           *sql.DB
	BackupDir    string
	ContentDir   string
	StaticDir    string
	TinaDataDir  string
	MaxBackups   int
	IncludeMedia bool
}

// Backup performs a full backup of the database and content
func Backup(config BackupConfig) (string, error) {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return "", err
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("backup_%s", timestamp)
	backupDir := filepath.Join(config.BackupDir, backupName)
	
	// Create backup directory
	if err := os.Mkdir(backupDir, 0755); err != nil {
		return "", err
	}

	// Backup database
	if err := backupDatabase(config.DB, backupDir); err != nil {
		return "", err
	}

	// Backup content
	if err := backupContent(config.ContentDir, backupDir); err != nil {
		return "", err
	}

	// Backup TinaCMS data
	if err := backupTinaData(config.TinaDataDir, backupDir); err != nil {
		return "", err
	}

	// Backup static files (optional)
	if config.IncludeMedia {
		if err := backupStaticFiles(config.StaticDir, backupDir); err != nil {
			return "", err
		}
	}

	// Clean up old backups
	if config.MaxBackups > 0 {
		if err := cleanOldBackups(config.BackupDir, config.MaxBackups); err != nil {
			return "", err
		}
	}

	return backupDir, nil
}

// Restore restores data from a backup
func Restore(config BackupConfig, backupDir string) error {
	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	// Restore database
	if err := restoreDatabase(config.DB, backupDir); err != nil {
		return err
	}

	// Restore content
	if err := restoreContent(config.ContentDir, backupDir); err != nil {
		return err
	}

	// Restore TinaCMS data
	if err := restoreTinaData(config.TinaDataDir, backupDir); err != nil {
		return err
	}

	// Restore static files (optional)
	if config.IncludeMedia {
		if err := restoreStaticFiles(config.StaticDir, backupDir); err != nil {
			return err
		}
	}

	return nil
}

// restoreDatabase restores the database from a backup
func restoreDatabase(db *sql.DB, backupDir string) error {
	// Check if backup file exists
	backupFile := filepath.Join(backupDir, "database.sql")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("database backup file not found: %s", backupFile)
	}
	
	// Get database connection info
	var (
		dbname string
		user   string
		host   string
		port   string
	)
	
	err := db.QueryRow("SELECT current_database(), current_user, inet_server_addr(), inet_server_port()").Scan(&dbname, &user, &host, &port)
	if err != nil {
		return err
	}
	
	// Use psql to restore backup
	cmd := exec.Command("psql", "-h", host, "-p", port, "-U", user, "-d", dbname, "-f", backupFile)
	
	// Set PGPASSWORD environment variable from config
	if pgPassword := os.Getenv("DB_PASSWORD"); pgPassword != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", pgPassword))
	}
	
	// Run command
	if err := cmd.Run(); err != nil {
		// Fallback to manual SQL restore if psql is not available
		return sqlRestore(db, backupFile)
	}
	
	return nil
}

// sqlRestore restores the database from a SQL file
func sqlRestore(db *sql.DB, backupFile string) error {
	// Read SQL file
	sqlData, err := ioutil.ReadFile(backupFile)
	if err != nil {
		return err
	}
	
	// Split by semicolon to get individual statements
	// Note: This is a simple approach and may not work for all SQL statements
	statements := strings.Split(string(sqlData), ";")
	
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue // Skip empty lines and comments
		}
		
		if _, err := tx.Exec(stmt); err != nil {
			tx.Rollback()
			return err
		}
	}
	
	// Commit transaction
	return tx.Commit()
}

// backupContent backs up Hugo content files
func backupContent(contentDir string, backupDir string) error {
	// Create destination directory
	destDir := filepath.Join(backupDir, "content")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	
	// Copy content directory recursively
	return copyDir(contentDir, destDir)
}

// restoreContent restores Hugo content files
func restoreContent(contentDir string, backupDir string) error {
	// Check if backup directory exists
	srcDir := filepath.Join(backupDir, "content")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("content backup directory not found: %s", srcDir)
	}
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return err
	}
	
	// Clear destination directory
	if err := clearDir(contentDir); err != nil {
		return err
	}
	
	// Copy content directory recursively
	return copyDir(srcDir, contentDir)
}

// backupTinaData backs up TinaCMS data
func backupTinaData(tinaDir string, backupDir string) error {
	// Create destination directory
	destDir := filepath.Join(backupDir, "tina")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	
	// Copy TinaCMS directory recursively
	return copyDir(tinaDir, destDir)
}

// restoreTinaData restores TinaCMS data
func restoreTinaData(tinaDir string, backupDir string) error {
	// Check if backup directory exists
	srcDir := filepath.Join(backupDir, "tina")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("TinaCMS backup directory not found: %s", srcDir)
	}
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(tinaDir, 0755); err != nil {
		return err
	}
	
	// Clear destination directory
	if err := clearDir(tinaDir); err != nil {
		return err
	}
	
	// Copy TinaCMS directory recursively
	return copyDir(srcDir, tinaDir)
}

// backupStaticFiles backs up static files (media)
func backupStaticFiles(staticDir string, backupDir string) error {
	// Create destination directory
	destDir := filepath.Join(backupDir, "static")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	
	// Copy static directory recursively
	return copyDir(staticDir, destDir)
}

// restoreStaticFiles restores static files (media)
func restoreStaticFiles(staticDir string, backupDir string) error {
	// Check if backup directory exists
	srcDir := filepath.Join(backupDir, "static")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("static files backup directory not found: %s", srcDir)
	}
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return err
	}
	
	// Copy static directory recursively
	return copyDir(srcDir, staticDir)
}

// cleanOldBackups removes old backups, keeping only the specified number
func cleanOldBackups(backupDir string, maxBackups int) error {
	// Get list of backups
	backups, err := ListBackups(backupDir)
	if err != nil {
		return err
	}
	
	// Sort backups by timestamp (newer first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})
	
	// Remove old backups
	if len(backups) > maxBackups {
		for _, backup := range backups[maxBackups:] {
			backupPath := filepath.Join(backupDir, backup)
			if err := os.RemoveAll(backupPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	// Get file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	// Read source directory
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	
	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			// Recursively copy directory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// copyFile copies a file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	// Get file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	
	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	
	// Set permissions
	return os.Chmod(dst, srcInfo.Mode())
}

// clearDir removes all contents of a directory but keeps the directory itself
func clearDir(dir string) error {
	// Read directory
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	
	// Remove each entry
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	
	return nil
}

// ListBackups returns a list of available backups
func ListBackups(backupDir string) ([]string, error) {
	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil, nil // Return empty list if directory doesn't exist
	}

	// Read backup directory
	entries, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	// Filter directories that match backup pattern
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 7 && entry.Name()[:7] == "backup_" {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// backupDatabase backs up the database using pg_dump
func backupDatabase(db *sql.DB, backupDir string) error {
	// Get database connection info
	var (
		dbname string
		user   string
		host   string
		port   string
	)
	
	err := db.QueryRow("SELECT current_database(), current_user, inet_server_addr(), inet_server_port()").Scan(&dbname, &user, &host, &port)
	if err != nil {
		return err
	}

	// Create backup file
	backupFile := filepath.Join(backupDir, "database.sql")
	
	// Use pg_dump to create backup
	cmd := exec.Command("pg_dump", "-h", host, "-p", port, "-U", user, "-d", dbname, "-f", backupFile)
	
	// Set PGPASSWORD environment variable from config
	// For security, in a real-world scenario, use a .pgpass file or connection string
	if pgPassword := os.Getenv("DB_PASSWORD"); pgPassword != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", pgPassword))
	}
	
	// Run command
	if err := cmd.Run(); err != nil {
		// Fallback to SQL dump if pg_dump is not available
		return sqlDump(db, backupFile)
	}
	
	return nil
}

// sqlDump creates a SQL dump of the database using queries
func sqlDump(db *sql.DB, backupFile string) error {
	// Get list of tables
	rows, err := db.Query(`
		SELECT tablename FROM pg_catalog.pg_tables 
		WHERE schemaname = 'public'
		ORDER BY tablename
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	// Open backup file
	file, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Add header
	file.WriteString("-- SQL Dump\n")
	file.WriteString("-- Generated on " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	
	// Process each table
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}
	
	// First, dump schema
	for _, tableName := range tables {
		// Get table schema
		var schemaSQL bytes.Buffer
		schemaSQL.WriteString(fmt.Sprintf("-- Table: %s\n", tableName))
		schemaSQL.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))
		
		// Get column information
		colRows, err := db.Query(`
			SELECT 
				column_name, 
				data_type, 
				is_nullable, 
				column_default
			FROM information_schema.columns 
			WHERE table_name = $1
			ORDER BY ordinal_position
		`, tableName)
		if err != nil {
			return err
		}
		defer colRows.Close()
		
		// Process columns
		var columns []string
		for colRows.Next() {
			var colName, dataType, isNullable string
			var colDefault sql.NullString
			
			if err := colRows.Scan(&colName, &dataType, &isNullable, &colDefault); err != nil {
				return err
			}
			
			// Build column definition
			colDef := fmt.Sprintf("  %s %s", colName, dataType)
			
			// Add NOT NULL if needed
			if isNullable == "NO" {
				colDef += " NOT NULL"
			}
			
			// Add DEFAULT if set
			if colDefault.Valid {
				colDef += fmt.Sprintf(" DEFAULT %s", colDefault.String)
			}
			
			columns = append(columns, colDef)
		}
		
		// Join columns and finalize schema
		schemaSQL.WriteString(strings.Join(columns, ",\n"))
		schemaSQL.WriteString("\n);\n\n")
		
		// Write schema to file
		file.WriteString(schemaSQL.String())
	}
	
	// Then, dump data for each table
	for _, tableName := range tables {
		// Write table header
		file.WriteString(fmt.Sprintf("-- Data for table: %s\n", tableName))
		
		// Get all data from table
		dataRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
		if err != nil {
			return err
		}
		defer dataRows.Close()
		
		// Get column info
		columns, err := dataRows.Columns()
		if err != nil {
			return err
		}
		
		// Prepare for scanning
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		// Process each row
		for dataRows.Next() {
			if err := dataRows.Scan(valuePtrs...); err != nil {
				return err
			}
			
			// Build INSERT statement
			file.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
			file.WriteString(strings.Join(columns, ", "))
			file.WriteString(") VALUES (")
			
			// Format values
			valueStrs := make([]string, len(columns))
			for i, val := range values {
				if val == nil {
					valueStrs[i] = "NULL"
				} else {
					switch v := val.(type) {
					case []byte:
						valueStrs[i] = fmt.Sprintf("'%s'", string(v))
					case string:
						valueStrs[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
					case time.Time:
						valueStrs[i] = fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
					default:
						valueStrs[i] = fmt.Sprintf("%v", v)
					}
				}
			}
			
			file.WriteString(strings.Join(valueStrs, ", "))
			file.WriteString(");\n")
		}
		
		file.WriteString("\n")
	}
	
	return nil
}