package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ammar1510/converse/internal/api"
	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/database"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Set Gin mode based on environment
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize JWT key from environment variable
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	auth.InitJWTKey([]byte(jwtSecret))

	// Determine database type from environment (default to PostgreSQL)
	dbTypeStr := os.Getenv("DB_TYPE")
	if dbTypeStr == "" {
		dbTypeStr = "postgres" // Default to PostgreSQL
	}

	dbType := database.DatabaseType(dbTypeStr)

	// Get connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to individual connection parameters if DATABASE_URL not set
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")

		if dbHost == "" || dbName == "" || dbUser == "" {
			log.Fatal("Database connection details missing. Set DATABASE_URL or individual DB_* variables")
		}

		// Build connection string based on database type
		switch dbType {
		case database.PostgreSQL:
			dbURL = fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s?sslmode=disable",
				dbUser, dbPass, dbHost, dbPort, dbName,
			)
		case database.MySQL:
			dbURL = fmt.Sprintf(
				"mysql://%s:%s@tcp(%s:%s)/%s",
				dbUser, dbPass, dbHost, dbPort, dbName,
			)
		}
	}

	// Create database connection using factory
	db, err := database.NewDatabase(dbType, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("Connected to %s database successfully", dbType)

	// Initialize router with default middleware (logger and recovery)
	router := gin.Default()

	// Configure CORS using environment variable
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create API handlers
	authHandler := api.NewAuthHandler(db)
	messageHandler := api.NewMessageHandler(db)

	// Set up API routes
	// Public routes (no authentication required)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)

	// Protected routes (authentication required)
	authorized := router.Group("/api")
	authorized.Use(api.AuthMiddleware())
	{
		authorized.GET("/auth/me", authHandler.GetMe)

		// Message routes
		authorized.POST("/messages", messageHandler.SendMessage)
		authorized.GET("/messages", messageHandler.GetMessages)
		authorized.GET("/messages/conversation/:userID", messageHandler.GetConversation)
		authorized.PUT("/messages/:messageID/read", messageHandler.MarkMessageAsRead)

		// More protected routes can be added here
	}

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Get server port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server 5 seconds to finish processing remaining requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
