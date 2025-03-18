package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"github.com/ammar1510/converse/internal/api"
	"github.com/ammar1510/converse/internal/auth"
	"github.com/ammar1510/converse/internal/database"
	internalWs "github.com/ammar1510/converse/internal/websocket"
)

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Configure log to write to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// Add timestamps to log entries
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Server logging initialized - output directed to console and server.log")

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

	// Initialize WebSocket manager
	wsManager := internalWs.NewManager()
	go wsManager.Run()

	// Set the WebSocket manager in the messages package
	api.WSManager = wsManager

	// Set up API routes
	// Public routes (no authentication required)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)

	// Protected routes (authentication required)
	authorized := router.Group("/api")
	authorized.Use(api.AuthMiddleware())
	{
		authorized.GET("/auth/me", authHandler.GetMe)
		authorized.GET("/users", authHandler.GetAllUsers)

		// Message routes
		authorized.POST("/messages", messageHandler.SendMessage)
		authorized.GET("/messages", messageHandler.GetMessages)
		authorized.GET("/messages/conversation/:userID", messageHandler.GetConversation)
		authorized.PUT("/messages/:messageID/read", messageHandler.MarkMessageAsRead)

		// WebSocket route with special middleware for token in URL parameter
		authorized.GET("/ws", func(c *gin.Context) {
			remoteAddr := c.Request.RemoteAddr
			log.Printf("[WebSocket] Connection request received from %s", remoteAddr)

			// Check for token in URL parameter
			tokenParam := c.Query("token")
			if tokenParam != "" {
				tokenPreview := tokenParam
				if len(tokenParam) > 10 {
					tokenPreview = tokenParam[:10] + "..." // Show only first 10 chars for security
				}
				log.Printf("[WebSocket] Found token in URL parameter from %s: %s", remoteAddr, tokenPreview)

				// Validate token from query parameter
				claims, err := auth.ValidateToken(tokenParam)
				if err == nil {
					log.Printf("[WebSocket] Token validated successfully for %s", remoteAddr)
					// Parse user ID string into UUID
					if userUUID, err := uuid.Parse(claims.UserID); err == nil {
						log.Printf("[WebSocket] User authenticated from URL parameter: %s (IP: %s)", userUUID, remoteAddr)
						// Set user ID and username in context
						c.Set("userID", userUUID)
						c.Set("username", claims.Username)
						// Continue to WebSocket handler
						wsManager.HandleWebSocket(c)
						return
					} else {
						log.Printf("[WebSocket] Failed to parse user ID from token for %s: %v", remoteAddr, err)
					}
				} else {
					log.Printf("[WebSocket] Token validation failed for %s: %v", remoteAddr, err)
				}
			} else {
				log.Printf("[WebSocket] No token found in URL parameter for %s", remoteAddr)
			}

			// If no valid token in URL parameter, try normal auth middleware
			// This will check for Authorization header
			if _, exists := c.Get("userID"); exists {
				log.Printf("User already authenticated by middleware")
				// User is already authenticated by the auth middleware
				wsManager.HandleWebSocket(c)
				return
			}

			// No valid authentication found
			log.Printf("No valid authentication found, returning unauthorized")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		})

		// More protected routes can be added here
	}

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Add test log endpoint
	router.GET("/test-log", func(c *gin.Context) {
		log.Println("Test log entry - this should appear in server.log")
		c.JSON(http.StatusOK, gin.H{"message": "Log test triggered"})
	})

	// Public WebSocket test endpoint (no auth required)
	router.GET("/ws-public", func(c *gin.Context) {
		fmt.Println("==== Public WebSocket endpoint hit ====")
		fmt.Printf("Request from: %s\n", c.Request.RemoteAddr)
		fmt.Printf("Method: %s\n", c.Request.Method)
		fmt.Printf("Protocol: %s\n", c.Request.Proto)

		// Log all headers
		fmt.Println("Headers:")
		for name, values := range c.Request.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}

		// Check if it's a WebSocket upgrade request
		if c.GetHeader("Upgrade") != "websocket" {
			fmt.Println("Non-WebSocket request received")
			c.String(http.StatusOK, "WebSocket endpoint is working, but you need to connect with a WebSocket client")
			return
		}

		// Upgrade HTTP connection to WebSocket
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				fmt.Printf("Checking origin: %s\n", r.Header.Get("Origin"))
				return true // Allow all origins
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade connection: %v\n", err)
			return
		}

		fmt.Println("Connection upgraded successfully")

		// Send a test message
		err = conn.WriteMessage(websocket.TextMessage, []byte("Hello from server"))
		if err != nil {
			fmt.Printf("Error writing message: %v\n", err)
		}

		// Keep the connection open for a while
		time.Sleep(5 * time.Second)

		// Close after sending
		conn.Close()
		fmt.Println("Connection closed")
	})

	// Root WebSocket endpoint
	router.GET("/socket", func(c *gin.Context) {
		fmt.Println("==== Root WebSocket endpoint hit ====")

		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade connection: %v\n", err)
			return
		}

		fmt.Println("Root socket connection upgraded successfully")
		conn.WriteMessage(websocket.TextMessage, []byte("Hello from root socket"))
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
