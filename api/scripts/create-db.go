package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/radjathaher/alunalun/api/internal/config"
)

var (
	yesFlag = flag.Bool("yes", false, "Skip confirmation prompts")
)

func main() {
	flag.Parse()
	
	log.Println("🏗️  Creating Development Database")
	log.Println("=================================")

	// Load configuration
	cfg := config.Load()
	
	// Build dev database URL for display
	devDBName := "dev_alunalun"
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, devDBName, cfg.DB.SSLMode)

	// Show target database
	log.Printf("🎯 Target Database: %s", dbURL)
	log.Printf("📍 Database Name: %s", devDBName)
	
	// Confirmation prompt (unless -yes flag)
	if !*yesFlag {
		if !confirmAction() {
			log.Println("❌ Operation cancelled")
			os.Exit(0)
		}
	}

	// Create database
	if err := createDatabaseIfNotExists(cfg.DB, devDBName); err != nil {
		log.Fatalf("❌ Database creation failed: %v", err)
	}

	log.Println("✅ Database creation completed successfully!")
}

func confirmAction() bool {
	fmt.Print("📝 Proceed with database creation? (Y/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "" || response == "y" || response == "yes"
}

func createDatabaseIfNotExists(dbConfig config.DBConfig, devDBName string) error {
	log.Println("🏗️  Creating database if needed...")
	
	ctx := context.Background()
	
	// Connect to postgres database to create dev database if needed
	adminURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.SSLMode)
	
	adminConn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminConn.Close(ctx)

	// Check if database exists
	var exists bool
	err = adminConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", devDBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		log.Printf("📋 Database '%s' already exists", devDBName)
		return nil
	}

	// Create database
	if _, err := adminConn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, devDBName)); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("🆕 Created database: %s", devDBName)
	return nil
}