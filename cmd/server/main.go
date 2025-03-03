package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	
	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/database"
	"github.com/ammar1510/converse/internal/api"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}
	
	// Verify JWT secret is set
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	
	// Set application mode based on environment
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Set up database connection
	dbConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("SUPABASE_DB_HOST"),
		os.Getenv("SUPABASE_DB_PORT"),
		os.Getenv("SUPABASE_DB_USER"),
		os.Getenv("SUPABASE_DB_PASSWORD"),
		os.Getenv("SUPABASE_DB_NAME"),
	)
	
	db, err := database.Connect(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Set up router
	router := gin.Default()
	
	// Set up routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		v1.POST("/auth/register", handlers.Register(db))
		v1.POST("/auth/login", handlers.Login(db))
		
		// Protected routes
		authorized := v1.Group("/")
		authorized.Use(auth.JWTMiddleware())
		{
			// User routes
			authorized.GET("/user/profile", handlers.GetUserProfile(db))
			// Will add chat routes here later
		}
	}
	
	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 