package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/naufalrafianto/lynx-api/internal/config"
	"github.com/naufalrafianto/lynx-api/internal/handlers"
	"github.com/naufalrafianto/lynx-api/internal/interfaces"
	"github.com/naufalrafianto/lynx-api/internal/middleware"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/services"
	"github.com/naufalrafianto/lynx-api/internal/utils"
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
	app.Run()
}

func (a *App) Initialize() error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.config = cfg

	// Initialize logger
	utils.InitLogger(cfg.AppEnv)

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

	return nil
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.config.Port),
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

	// Middlewares
	router.Use(gin.Recovery())
	router.Use(utils.NewLoggerMiddleware(utils.Logger).Handle())
	router.Use(cors.New(a.corsConfig()))

	// Determine base URL
	baseURL := fmt.Sprintf("http://%s:%s", a.config.Host, a.config.Port)
	if a.config.AppEnv == "production" && a.config.BaseURL != "" {
		baseURL = a.config.BaseURL
	}

	// Initialize services with interfaces
	var authService interfaces.AuthService = services.NewAuthService(a.db, a.redis)
	var urlService interfaces.URLService = services.NewURLService(a.db, a.redis, a.config.URLPrefix)
	var qrService interfaces.QRService = services.NewQRService(a.db, a.redis, baseURL)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, a.config.JWTSecret)
	urlHandler := handlers.NewURLHandler(urlService, baseURL)
	qrHandler := handlers.NewQRHandler(qrService, urlService)

	// Health check
	router.GET("/health", a.healthCheck())

	// Public routes
	router.GET("/qr/:shortCode", qrHandler.GetQRCode)
	router.GET("/qr/:shortCode/base64", qrHandler.GetQRCodeBase64)
	router.GET("/urls/:shortCode", urlHandler.RedirectToLongURL)

	// API routes
	v1 := router.Group("/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		api := v1.Group("/api")
		api.Use(middleware.AuthMiddleware(a.config.JWTSecret))
		{
			// User routes
			user := api.Group("/user")
			{
				user.GET("/me", authHandler.GetUserDetails)
				user.POST("/logout", authHandler.Logout)
			}

			// URL routes
			urls := api.Group("/urls")
			{
				urls.POST("", urlHandler.CreateShortURL)
				urls.GET("", urlHandler.GetUserURLs)
				urls.GET("/:id", urlHandler.GetURL)
				urls.DELETE("/:id", urlHandler.DeleteURL)
			}
		}
	}

	router.NoRoute(a.notFound())

	return router
}
func (a *App) corsConfig() cors.Config {
	return cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		a.config.DBHost, a.config.DBUser, a.config.DBPassword, a.config.DBName, a.config.DBPort)

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
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return redisClient, nil
}

func (a *App) initMigrations() error {
	return a.db.AutoMigrate(
		&models.User{},
		&models.URL{},
	)
}
