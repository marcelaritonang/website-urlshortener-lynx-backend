package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/config"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🔧 Starting Database Migration Test...")
	fmt.Println(strings.Repeat("=", 50))

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// Display connection info
	fmt.Println("\n📊 Database Configuration:")
	fmt.Printf("  Host:     %s\n", cfg.DBHost)
	fmt.Printf("  Port:     %s\n", cfg.DBPort)
	fmt.Printf("  User:     %s\n", cfg.DBUser)
	fmt.Printf("  Database: %s\n", cfg.DBName)
	fmt.Printf("  Password: %s\n", maskPassword(cfg.DBPassword))

	// Build DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	fmt.Println("\n🔌 Connecting to database...")

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	fmt.Println("✅ Database connection successful!")

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database instance:", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}

	fmt.Println("✅ Database ping successful!")

	// Run migrations
	fmt.Println("\n🚀 Running migrations...")

	if err := db.AutoMigrate(&models.User{}, &models.URL{}); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	fmt.Println("✅ Migrations completed successfully!")

	// Verify tables created
	fmt.Println("\n📋 Verifying tables...")

	var tables []struct {
		TableName string
	}

	db.Raw("SELECT tablename as table_name FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables)

	if len(tables) == 0 {
		log.Fatal("❌ No tables found!")
	}

	fmt.Println("✅ Tables created:")
	for _, table := range tables {
		fmt.Printf("  - %s\n", table.TableName)
	}

	// Check table structure
	fmt.Println("\n🔍 Checking table structures...")

	// Check users table
	var userColumns []struct {
		ColumnName string
		DataType   string
	}
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'users' 
		ORDER BY ordinal_position
	`).Scan(&userColumns)

	fmt.Println("\n✅ Users table columns:")
	for _, col := range userColumns {
		fmt.Printf("  - %-20s %s\n", col.ColumnName, col.DataType)
	}

	// Check urls table
	var urlColumns []struct {
		ColumnName string
		DataType   string
	}
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'urls' 
		ORDER BY ordinal_position
	`).Scan(&urlColumns)

	fmt.Println("\n✅ URLs table columns:")
	for _, col := range urlColumns {
		fmt.Printf("  - %-20s %s\n", col.ColumnName, col.DataType)
	}

	// Check indexes
	fmt.Println("\n🔑 Checking indexes...")

	var indexes []struct {
		TableName string
		IndexName string
	}
	db.Raw(`
		SELECT tablename as table_name, indexname as index_name 
		FROM pg_indexes 
		WHERE schemaname = 'public' 
		ORDER BY tablename, indexname
	`).Scan(&indexes)

	fmt.Println("✅ Indexes created:")
	currentTable := ""
	for _, idx := range indexes {
		if idx.TableName != currentTable {
			fmt.Printf("\n  %s:\n", idx.TableName)
			currentTable = idx.TableName
		}
		fmt.Printf("    - %s\n", idx.IndexName)
	}

	// Count existing data
	fmt.Println("\n📊 Checking existing data...")

	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	fmt.Printf("  Users:  %d records\n", userCount)

	var urlCount int64
	db.Model(&models.URL{}).Count(&urlCount)
	fmt.Printf("  URLs:   %d records\n", urlCount)

	// Database statistics
	fmt.Println("\n💾 Database statistics:")

	var dbSize string
	db.Raw("SELECT pg_size_pretty(pg_database_size(?)) as size", cfg.DBName).Scan(&dbSize)
	fmt.Printf("  Database size: %s\n", dbSize)

	// Connection pool stats
	stats := sqlDB.Stats()
	fmt.Println("\n🔌 Connection pool stats:")
	fmt.Printf("  Open connections:    %d\n", stats.OpenConnections)
	fmt.Printf("  In use:              %d\n", stats.InUse)
	fmt.Printf("  Idle:                %d\n", stats.Idle)

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 Database migration test completed successfully!")
	fmt.Println(strings.Repeat("=", 50) + "\n")
}

func maskPassword(password string) string {
	if len(password) == 0 {
		return "(empty)"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}
