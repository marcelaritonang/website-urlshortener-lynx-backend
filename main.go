package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/config"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/handlers"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/interfaces"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/middleware"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/services"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type App struct {
	config *config.Config
	db     *gorm.DB
	redis  *redis.Client
	router *gin.Engine
}

func main() {
	app := &App{}
	if err := app.Initialize(); err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// PENTING: Baca PORT dari environment variable (override config jika ada)
	port := os.Getenv("PORT")
	if port == "" {
		port = app.config.Port
	}
	app.config.Port = port

	// Baca DATABASE_URL dari environment (sudah di-handle di config.go)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		log.Printf("DATABASE_URL: %s", dbURL)
	}

	log.Printf("Starting server on port %s", port)
	app.Run()
}

func (a *App) Initialize() error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.config = cfg

	// ✅ FIX: Initialize logger FIRST (before using utils.Logger)
	utils.InitLogger(cfg.AppEnv)

	// ✅ NOW safe to use utils.Logger
	utils.Logger.Info("JWT Secret validated", "length", len(cfg.JWTSecret))

	// Initialize database
	db, err := a.initDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = db

	// Initialize Redis
	redis, err := a.initRedis()
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}
	a.redis = redis

	// Run migrations
	if err := a.initMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup router
	a.router = a.setupRouter()

	// ✅ NEW: Start cache warming service
	cacheWarmer := services.NewCacheWarmer(a.db, a.redis)
	cacheWarmer.StartCacheWarmer()

	return nil
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    ":" + a.config.Port,
		Handler: a.router,
	}

	// Graceful shutdown setup
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start server
	go func() {
		utils.Logger.Info("Server starting", "port", a.config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Error("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	utils.Logger.Info("Shutting down server...")

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		utils.Logger.Error("Server forced to shutdown", "error", err)
	}

	if err := a.redis.Close(); err != nil {
		utils.Logger.Error("Error closing Redis connection", "error", err)
	}

	utils.Logger.Info("Server exited properly")
}

func (a *App) setupRouter() *gin.Engine {
	if a.config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// ============================================================
	// CRITICAL: CUSTOM CORS MIDDLEWARE MUST BE FIRST!
	// ============================================================
	router.Use(middleware.CORSMiddleware())

	// Middleware lain SETELAH CORS
	router.Use(gin.Recovery())
	router.Use(utils.NewLoggerMiddleware(utils.Logger).Handle())
	router.Use(middleware.RateLimiterMiddleware(a.redis, middleware.RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         20,
		BlockDuration:     30 * time.Minute,
	}))

	// Determine base URL
	baseURL := fmt.Sprintf("http://%s:%s", a.config.Host, a.config.Port)
	if a.config.AppEnv == "production" && a.config.BaseURL != "" {
		baseURL = a.config.BaseURL
	}

	// ✅ Initialize services with interfaces
	var authService interfaces.AuthService = services.NewAuthService(a.db, a.redis)
	var urlService interfaces.URLService = services.NewURLService(a.db, a.redis, a.config.URLPrefix)
	var qrService interfaces.QRService = services.NewQRService(a.db, a.redis, baseURL)

	// ✅ Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, a.config.JWTSecret, a.db)
	urlHandler := handlers.NewURLHandler(urlService, baseURL)
	qrHandler := handlers.NewQRHandler(qrService, urlService)

	// ============================================================
	// PUBLIC ROUTES (No Authentication)
	// ============================================================

	// Health check
	router.GET("/health", a.healthCheck())

	// QR Code generation
	router.GET("/qr/:shortCode", qrHandler.GetQRCode)
	router.GET("/qr/:shortCode/base64", qrHandler.GetQRCodeBase64)

	// URL Redirect
	router.GET("/urls/:shortCode", urlHandler.RedirectToLongURL)

	fmt.Println("✅ [ROUTER] Redirect route registered: GET /urls/:shortCode")
	fmt.Println("🔧 [ROUTER] Registering public routes...")

	// Public API routes (no authentication required)
	publicAPI := router.Group("/api")
	{
		publicAPI.POST("/urls", urlHandler.CreateAnonymousURL)
	}

	// ============================================================
	// API v1 ROUTES
	// ============================================================
	v1 := router.Group("/v1")
	{
		// Auth routes (public) - WITH STRICT RATE LIMITING
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimiterMiddleware(a.redis))
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password",
				middleware.ForgotPasswordRateLimiter(a.redis),
				authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPasswordConfirm)
		}

		// Protected routes (authentication required)
		api := v1.Group("/api")
		api.Use(middleware.AuthMiddleware(a.config.JWTSecret))
		{
			// User routes
			user := api.Group("/user")
			{
				user.GET("/me", authHandler.GetUserDetails)
				user.POST("/logout", authHandler.Logout)
			}

			// URL routes (authenticated users only)
			urls := api.Group("/urls")
			{
				urls.POST("", urlHandler.CreateShortURL)
				urls.GET("", urlHandler.GetUserURLs)
				urls.GET("/:id", urlHandler.GetURL)
				urls.DELETE("/:id", urlHandler.DeleteURL)
			}
		}
	}

	// 404 handler
	router.NoRoute(a.notFound())

	return router
}

func (a *App) healthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, "Service is healthy", gin.H{
			"time": time.Now().UTC(),
		})
	}
}

func (a *App) notFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotFound, errors.New("route not found"))
	}
}

func (a *App) initDatabase() (*gorm.DB, error) {
	fmt.Println("=== DATABASE CONNECTION DEBUG ===")
	fmt.Println("DBHost:", a.config.DBHost)
	fmt.Println("DBPort:", a.config.DBPort)
	fmt.Println("DBUser:", a.config.DBUser)
	fmt.Println("DBPassword:", a.config.DBPassword)
	fmt.Println("DBName:", a.config.DBName)

	// ✅ Render requires sslmode=require
	sslMode := "disable"
	if a.config.AppEnv == "production" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		a.config.DBHost, a.config.DBUser, a.config.DBPassword, a.config.DBName, a.config.DBPort, sslMode)

	// Don't print password in production
	if a.config.AppEnv != "production" {
		fmt.Println("DSN:", dsn)
	}
	fmt.Println("================================")

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	if a.config.AppEnv == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	return gorm.Open(postgres.Open(dsn), gormConfig)
}

func (a *App) initRedis() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", a.config.RedisHost, a.config.RedisPort),
		Password:     a.config.RedisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Test connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	// ✅ Test write operation
	testKey := "test:connection"
	if err := redisClient.Set(ctx, testKey, "ok", 10*time.Second).Err(); err != nil {
		return nil, fmt.Errorf("redis write test failed: %w", err)
	}

	// ✅ Test read operation
	if val, err := redisClient.Get(ctx, testKey).Result(); err != nil || val != "ok" {
		return nil, fmt.Errorf("redis read test failed")
	}

	// Cleanup test key
	redisClient.Del(ctx, testKey)

	fmt.Println("✅ Redis connection tested successfully")

	return redisClient, nil
}

func (a *App) initMigrations() error {
	fmt.Println("🔄 Running database migrations...")

	// ✅ Configure connection pool
	sqlDB, err := a.db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ✅ Run migrations
	if err := a.db.AutoMigrate(
		&models.User{},
		&models.URL{},
	); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// ✅ Verify tables exist
	var tableCount int64
	if err := a.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'urls')").Scan(&tableCount).Error; err != nil {
		return fmt.Errorf("table verification failed: %w", err)
	}

	if tableCount != 2 {
		utils.Logger.Warn("Table verification", "expected", 2, "found", tableCount)

		// List existing tables
		var tables []string
		a.db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)
		utils.Logger.Info("Existing tables", "tables", tables)

		return fmt.Errorf("migration incomplete: expected 2 tables, found %d", tableCount)
	}

	utils.Logger.Info("Tables verified successfully", "count", tableCount)

	// ✅ Test database connection
	if sqlDB, err := a.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database ping failed after migration: %w", err)
		}
	}

	fmt.Println("✅ Migrations completed successfully")
	return nil
}
