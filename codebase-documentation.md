# Converse: Go Chat Service Documentation

## Overview

Converse is a high-performance, real-time chat service built with Go. This documentation provides a comprehensive overview of the codebase structure and functionality.

## Codebase Structure

```
converse/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── auth.go              # Authentication API handlers
│   │   └── middleware.go        # API middleware (JWT auth)
│   ├── auth/
│   │   ├── jwt.go               # JWT token handling
│   │   └── password.go          # Password hashing utilities
│   ├── database/
│   │   └── supabase.go          # Database connection and operations
│   └── models/
│       └── user.go              # Data models and validation
└── .env                         # Environment configuration
```

## Package: cmd/server

### main.go
Main application entrypoint that initializes and wires together all components.

- **`main()`**: Application entrypoint that:
  - Loads environment variables from `.env`
  - Initializes the JWT authentication system
  - Establishes database connection
  - Sets up API routes and middleware
  - Starts the HTTP server

## Package: internal/api

### auth.go
Implements authentication-related API endpoints.

- **`type AuthHandler struct`**: Handler with database reference for auth operations
- **`NewAuthHandler(db *database.DB)`**: Constructor for AuthHandler
- **`Register(c *gin.Context)`**: Handles user registration requests
  - Validates registration input
  - Hashes password
  - Creates user in database
  - Returns success/error response
- **`Login(c *gin.Context)`**: Handles user login requests
  - Verifies credentials
  - Generates JWT token
  - Updates user's last seen timestamp
  - Returns token with user data
- **`GetMe(c *gin.Context)`**: Retrieves authenticated user's profile
  - Gets user ID from JWT context
  - Returns user profile data

### middleware.go
Provides middleware for request preprocessing.

- **`AuthMiddleware()`**: JWT verification middleware
  - Extracts JWT token from Authorization header
  - Validates token authenticity
  - Sets user ID and username in request context
  - Aborts with 401 if token is invalid

## Package: internal/auth

### jwt.go
Handles JWT token generation and validation.

- **`type JWTClaims struct`**: Defines JWT payload structure
- **`GenerateToken(user *models.User)`**: Creates JWT token for authenticated user
  - Sets user ID and username in claims
  - Sets token expiration time
  - Signs token with secret key
- **`ValidateToken(tokenString string)`**: Verifies JWT token authenticity
  - Parses and validates token
  - Returns extracted claims or error
- **`GetUserIDFromToken(claims *JWTClaims)`**: Extracts UUID from token claims
- **`InitJWTKey(key []byte)`**: Sets the JWT secret key (to be implemented)

### password.go
Provides password security utilities.

- **`HashPassword(password string)`**: Securely hashes passwords using bcrypt
- **`CheckPasswordHash(password, hash string)`**: Verifies password against stored hash

## Package: internal/database

### supabase.go
Manages database connectivity and operations.

- **`type DB struct`**: Wrapper around SQL database connection
- **`NewDB(connStr string)`**: Establishes connection to Supabase Postgres
- **`CreateUser(username, email, passwordHash string)`**: Stores new user
  - Checks if user already exists
  - Creates user with UUID
  - Returns user object or error
- **`GetUserByEmail(email string)`**: Retrieves user by email address
  - Queries database for matching email
  - Returns user object or not found error
- **`UpdateLastSeen(userID uuid.UUID)`**: Updates user's last activity timestamp

## Package: internal/models

### user.go
Defines data structures and validation for user-related operations.

- **`type User struct`**: Complete user model with all fields
- **`type UserRegistration struct`**: Input validation for registration
- **`type UserLogin struct`**: Input validation for login
- **`type UserResponse struct`**: Public user data (excludes sensitive fields)

## Configuration

### .env
Environment configuration file.

- **JWT_SECRET**: Secret key for signing and verifying JWT tokens
- **SUPABASE_DB_HOST**: Supabase database hostname
- **SUPABASE_DB_PORT**: Database port (usually 5432)
- **SUPABASE_DB_NAME**: Database name (usually postgres)
- **SUPABASE_DB_USER**: Database username
- **SUPABASE_DB_PASSWORD**: Database password
- **REDIS_URL**: URL for Redis connection (not yet implemented)
- **PORT**: Port for HTTP server (default 8080)
- **ENV**: Application environment (development, production)

## Planned Features (Not Yet Implemented)

1. **WebSocket Handler**: For real-time chat communication
2. **Redis Integration**: For user presence and temporary data
3. **Chat Models and Operations**: For message persistence
4. **Channel Management**: For group conversations

## Authentication Flow

1. **Registration**: Client sends username, email, password → Server creates user → Returns user data
2. **Login**: Client sends email, password → Server verifies → Returns JWT token and user data
3. **Authenticated Requests**: Client includes JWT in Authorization header → Middleware validates → Handler processes request 