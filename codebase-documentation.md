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
│   │   ├── auth_test.go         # Tests for auth handlers
│   │   ├── middleware.go        # API middleware (JWT auth)
│   │   └── middleware_test.go   # Tests for auth middleware
│   ├── auth/
│   │   ├── jwt.go               # JWT token handling
│   │   ├── jwt_test.go          # Tests for JWT functionality
│   │   ├── password.go          # Password hashing utilities
│   │   └── password_test.go     # Tests for password functions
│   ├── database/
│   │   ├── supabase.go          # Database connection and operations
│   │   └── supabase_test.go     # Tests for database operations
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

### auth_test.go
Comprehensive tests for authentication endpoints.

- **`TestRegister`**: Tests user registration with various scenarios
  - Valid registration
  - Duplicate email/username
  - Invalid input validation
- **`TestLogin`**: Tests login functionality
  - Valid credentials
  - Invalid password
  - Non-existent user
  - Invalid input
- **`TestGetMe`**: Tests profile retrieval
  - Valid token
  - Missing/invalid token

### middleware.go
Provides middleware for request preprocessing.

- **`AuthMiddleware()`**: JWT verification middleware
  - Extracts JWT token from Authorization header
  - Validates token authenticity
  - Sets user ID and username in request context
  - Aborts with 401 if token is invalid

### middleware_test.go
Tests for authentication middleware.

- **`TestAuthMiddleware`**: Tests JWT validation
  - Valid token processing
  - Missing token
  - Invalid token format
  - Missing Bearer prefix

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
- **`InitJWTKey(key []byte)`**: Sets the JWT secret key for testing and initialization

### jwt_test.go
Tests for JWT functionality.

- **`TestInitJWTKey`**: Tests JWT key initialization
- **`TestGenerateToken`**: Tests token generation
  - Valid user
  - Missing user ID
  - Nil user
- **`TestValidateToken`**: Tests token validation
  - Valid token
  - Empty token
  - Invalid format
  - Tampered token
- **`TestGetUserIDFromToken`**: Tests user ID extraction
  - Valid claims
  - Invalid UUID format
  - Nil claims

### password.go
Provides password security utilities.

- **`HashPassword(password string)`**: Securely hashes passwords using bcrypt
- **`CheckPasswordHash(password, hash string)`**: Verifies password against stored hash

### password_test.go
Tests for password functionality.

- **`TestPasswordHashing`**: Tests password hashing
  - Common passwords
  - Empty passwords
  - Long passwords
  - Special characters
- **`TestCheckPasswordHash`**: Tests password verification
  - Correct password
  - Incorrect password
  - Empty password
  - Invalid hash

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

### supabase_test.go
Tests for database operations.

- **`TestNewDB`**: Tests database connection
  - Valid connection string
  - Invalid connection string
- **`TestCreateUser`**: Tests user creation
  - Valid user
  - Duplicate email
  - Duplicate username
- **`TestGetUserByEmail`**: Tests user retrieval
  - Existing user
  - Non-existent user
- **`TestUpdateLastSeen`**: Tests last seen update
  - Existing user
  - Non-existent user

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

## Testing Status

### Test Results Summary
1. **API Package Tests** (`internal/api/`)
   - ✅ `TestRegister`: All cases pass
     - Valid registration
     - Duplicate email handling
     - Input validation
   - ✅ `TestLogin`: All cases pass
     - Valid credentials
     - Invalid password
     - Non-existent user
     - Input validation
   - ✅ `TestAuthMiddleware`: All cases pass
     - Valid token
     - Missing token
     - Invalid format
     - Missing Bearer prefix
   - ⚠️ `TestGetMe`: Partially implemented
     - User profile retrieval needs database integration

2. **Database Package Tests** (`internal/database/`)
   - ✅ `TestNewDB`: Connection handling
   - ✅ `TestCreateUser`: User creation and validation
   - ✅ `TestGetUserByEmail`: User retrieval
   - ✅ `TestUpdateLastSeen`: Last activity tracking
   - ✅ `TestGetUserByID`: User lookup by UUID

3. **Auth Package Tests** (`internal/auth/`)
   - ✅ `TestPasswordHashing`: Password security
   - ✅ `TestCheckPasswordHash`: Hash verification
   - ✅ `TestGenerateToken`: JWT generation
   - ✅ `TestValidateToken`: JWT validation
   - ✅ `TestInitJWTKey`: JWT key management

### Environment Requirements
- PostgreSQL database with user-specific credentials
- Environment variables properly configured
- JWT secret key initialized

### Known Issues
1. Database tests require specific user configuration:
   - Using system username instead of 'postgres'
   - Proper database permissions
   - Test database (`converse_test`) created

### Planned Features (Not Yet Implemented)
1. **WebSocket Handler**: For real-time chat communication
2. **Redis Integration**: For user presence and temporary data
3. **Chat Models and Operations**: For message persistence
4. **Channel Management**: For group conversations

## Authentication Flow

1. **Registration**: Client sends username, email, password → Server creates user → Returns user data
2. **Login**: Client sends email, password → Server verifies → Returns JWT token and user data
3. **Authenticated Requests**: Client includes JWT in Authorization header → Middleware validates → Handler processes request 

## UI Implementation Plan

### Updated Minimal Implementation Approach

Based on development progress and prioritization, we've adopted a more focused, minimal implementation strategy to quickly get essential features working before expanding the UI.

#### Minimal Project Structure
```
converse-ui/
├── public/
│   ├── index.html
│   └── assets/
├── src/
│   ├── components/
│   │   ├── auth/                # Authentication components (priority)
│   │   │   ├── LoginForm.jsx
│   │   │   └── RegisterForm.jsx
│   │   ├── chat/                # Essential messaging components
│   │   │   ├── MessageList.jsx  # Display messages
│   │   │   └── MessageInput.jsx # Send messages
│   │   └── layout/              # Minimal layout
│   │       └── Navbar.jsx       # Simple navigation
│   ├── pages/                   # Core pages only
│   │   ├── LoginPage.jsx
│   │   ├── RegisterPage.jsx
│   │   └── ChatPage.jsx
│   ├── services/                # Minimal services
│   │   ├── authService.js       # Authentication API
│   │   └── messageService.js    # Message API
│   ├── context/                 # Essential context
│   │   └── AuthContext.jsx      # Authentication state
│   ├── App.jsx                  # Main application with routes
│   └── index.jsx                # Entry point
└── package.json
```

#### Phased Implementation Strategy

1. **Phase 1: Authentication (Current Focus)**
   - Implement login and registration functionality
   - Set up authentication context and token storage
   - Create protected routes
   - Connect to backend auth endpoints

2. **Phase 2: Basic Messaging**
   - Implement simple message list and input
   - Create message service for API communication
   - Add polling for new messages (no WebSockets yet)

3. **Phase 3: Enhanced Features**
   - Add real-time features as needed
   - Implement additional UI components from comprehensive plan
   - Improve styling and user experience

This minimal approach allows us to quickly get a functional UI that connects to our backend, while deferring more complex features until the core functionality is working properly.

### Comprehensive Frontend Architecture (Future Reference)
Converse will implement a React frontend with a separate application architecture, communicating with the Go backend via REST API and WebSockets.

### Project Structure
```
converse-ui/
├── public/
│   ├── index.html
│   ├── favicon.ico
│   └── assets/
├── src/
│   ├── components/
│   │   ├── auth/                # Authentication components
│   │   │   ├── LoginForm.jsx
│   │   │   └── RegisterForm.jsx
│   │   ├── chat/                # Messaging components
│   │   │   ├── MessageList.jsx
│   │   │   ├── MessageItem.jsx
│   │   │   ├── MessageInput.jsx
│   │   │   └── ConversationList.jsx
│   │   ├── layout/              # Layout components
│   │   │   ├── Header.jsx
│   │   │   ├── Sidebar.jsx
│   │   │   ├── Footer.jsx
│   │   │   └── Layout.jsx
│   │   ├── profile/             # User profile components
│   │   │   ├── UserProfile.jsx
│   │   │   └── ProfileSettings.jsx
│   │   └── common/              # Shared components
│   │       ├── Button.jsx
│   │       ├── Avatar.jsx
│   │       └── Loading.jsx
│   ├── pages/                   # Page components
│   │   ├── LoginPage.jsx
│   │   ├── RegisterPage.jsx
│   │   ├── ChatPage.jsx
│   │   ├── ProfilePage.jsx
│   │   └── NotFoundPage.jsx
│   ├── services/                # API and WebSocket services
│   │   ├── api.js               # REST API client
│   │   ├── websocket.js         # WebSocket client
│   │   ├── authService.js       # Authentication service
│   │   └── messageService.js    # Messaging service
│   ├── context/                 # React context providers
│   │   ├── AuthContext.jsx      # Authentication state
│   │   └── ChatContext.jsx      # Chat state
│   ├── hooks/                   # Custom React hooks
│   │   ├── useAuth.js           # Authentication hook
│   │   └── useChat.js           # Chat functionality hook
│   ├── utils/                   # Utility functions
│   │   ├── dateFormat.js        # Date formatting utilities
│   │   └── validators.js        # Input validation utilities
│   ├── styles/                  # CSS/SCSS files
│   │   ├── variables.scss       # Design tokens
│   │   └── global.scss          # Global styles
│   ├── App.jsx                  # Main application component
│   └── index.jsx                # Entry point
├── package.json
└── vite.config.js
```

### Key UI Layouts

#### Main Application Layout
```
┌─────────────────────────────────────────────────┐
│ Header (App Logo, User Status, Settings)        │
├─────────┬───────────────────────────────────────┤
│         │                                       │
│         │                                       │
│ Sidebar │        Main Content Area              │
│         │                                       │
│         │                                       │
├─────────┴───────────────────────────────────────┤
│ Footer (Optional - Version, Status)             │
└─────────────────────────────────────────────────┘
```

### Core Screens

1. **Authentication Screens**
   - Login Screen: Email/password form with validation
   - Registration Screen: Username, email, password fields

2. **Main Chat Interface**
   - Conversation list in sidebar (recent conversations)
   - Active conversation in main content area
   - Message input with typing indicator at bottom
   - User status indicators (online/offline)

3. **User Profile & Settings**
   - Profile information display
   - Account settings
   - Notification preferences
   - Security settings

### State Management

- **Authentication State**: JWT token, user profile, login status
- **Chat State**: Active conversations, message history, unread counts
- **UI State**: Active views, modals, responsive layout state
- **WebSocket State**: Connection status, real-time events

### User Flows

1. **Authentication Flow**
   - User enters credentials → Client validates → Server authenticates → Chat interface loads

2. **Messaging Flow**
   - User selects conversation → Message history loads → User sends message → Real-time updates

3. **Profile Management Flow**
   - User navigates to profile → Views/edits information → Saves changes → Server updates

### Integration with Backend

1. **REST API Integration**
   - Authentication endpoints (login, register)
   - Message retrieval (get conversations, message history)
   - Profile management (get/update profile)

2. **WebSocket Integration**
   - Real-time message delivery
   - Typing indicators
   - Online presence
   - Read receipts

### Responsive Design

- Mobile-first approach
- Breakpoints for different device sizes:
  - Small (mobile): 320px - 576px
  - Medium (tablet): 577px - 992px
  - Large (desktop): 993px+
- Collapsible sidebar on smaller screens
- Optimized layouts for different form factors

### Visual Design Elements

- **Color Palette**
  - Primary: #3498db (brand blue)
  - Secondary: #2ecc71 (success green)
  - Accent: #e74c3c (notification red)
  - Neutrals: #f8f9fa, #e9ecef, #dee2e6, #ced4da, #6c757d, #343a40

- **Typography**
  - Sans-serif system fonts for optimal performance
  - Clear hierarchy with defined text styles
  - Consistent font weights

- **Visual Elements**
  - Subtle shadows for elevation
  - Rounded corners for friendly feel
  - Clear read/unread indicators
  - User avatars and status indicators

### Implementation Phases

1. **Phase 1: Core Structure**
   - Project setup with React and Vite
   - Routing configuration
   - Authentication screens
   - Basic layout components

2. **Phase 2: Messaging UI**
   - Conversation list
   - Message view components
   - Message input
   - Basic styling

3. **Phase 3: Real-time Features**
   - REST API integration
   - WebSocket connection
   - Typing indicators
   - Message status indicators

4. **Phase 4: Polish & Refinement**
   - Animations and transitions
   - Performance optimization
   - Responsive design improvements
   - Error handling and fallbacks 