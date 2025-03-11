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
│   │   ├── messages.go          # Message API handlers
│   │   ├── messages_test.go     # Tests for message handlers
│   │   ├── middleware.go        # API middleware (JWT auth)
│   │   └── middleware_test.go   # Tests for auth middleware
│   ├── auth/
│   │   ├── jwt.go               # JWT token handling
│   │   ├── jwt_test.go          # Tests for JWT functionality
│   │   ├── password.go          # Password hashing utilities
│   │   └── password_test.go     # Tests for password functions
│   ├── database/
│   │   ├── database.go          # Database interface definition
│   │   ├── postgres.go          # PostgreSQL implementation
│   │   ├── postgres_test.go     # Tests for PostgreSQL operations
│   │   └── schema.sql           # Database schema definition
│   ├── models/
│   │   ├── user.go              # User data models and validation
│   │   └── message.go           # Message data models
│   └── websocket/
│       ├── handler.go           # WebSocket connection handling
│       └── handler_test.go      # Tests for WebSocket functionality
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
  - Initializes WebSocket manager
  - Starts the HTTP server
  - Implements graceful shutdown for clean termination
  - Configures CORS for cross-origin requests

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
  - Converts string ID to UUID
  - Fetches user data from database
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

### messages.go
Implements message-related API endpoints.

- **`type MessageHandler struct`**: Handler with database reference for message operations
- **`NewMessageHandler(db *database.DBInterface)`**: Constructor for MessageHandler
- **`SendMessage(c *gin.Context)`**: Handles message creation
  - Validates message input
  - Creates message in database
  - Notifies recipient via WebSocket if connected
  - Returns created message
- **`GetMessages(c *gin.Context)`**: Retrieves all messages for authenticated user
  - Gets user ID from JWT context
  - Fetches messages from database
  - Returns message list
- **`GetConversation(c *gin.Context)`**: Retrieves conversation between two users
  - Gets user ID from JWT context
  - Validates other user ID from URL parameter
  - Fetches conversation from database
  - Returns message list
- **`MarkMessageAsRead(c *gin.Context)`**: Marks a message as read
  - Gets user ID from JWT context
  - Validates message ID from URL parameter
  - Verifies user is the intended recipient
  - Updates message read status
  - Notifies sender via WebSocket

### messages_test.go
Comprehensive tests for message endpoints.

- **`TestCreateMessage`**: Tests message creation
  - Valid message creation
  - Invalid input
  - Unauthorized access
- **`TestGetMessages`**: Tests message retrieval
  - Valid retrieval
  - Empty message list
  - Unauthorized access
- **`TestGetConversation`**: Tests conversation retrieval
  - Valid conversation
  - Invalid user ID
  - Unauthorized access
- **`TestMarkMessageAsRead`**: Tests marking messages as read
  - Valid marking
  - Unauthorized access
  - Non-existent message
- **`TestWebSocketMessageFormat`**: Tests WebSocket message format

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

### database.go
Defines the database interface and factory for different implementations.

- **`type DBInterface interface`**: Interface that all database implementations must satisfy
  - User methods (CreateUser, GetUserByEmail, etc.)
  - Message methods (CreateMessage, GetMessagesByUser, etc.)
  - Common methods (Close, Exec)
- **`type DatabaseType string`**: Enum for supported database types
- **`NewDatabase(dbType DatabaseType, connStr string)`**: Factory function for database implementations

### postgres.go
PostgreSQL implementation of the database interface.

- **`type PostgresDB struct`**: PostgreSQL database implementation
- **`NewPostgresDB(connStr string)`**: Establishes connection to PostgreSQL
- **`CreateUser(username, email, passwordHash string)`**: Creates new user
- **`GetUserByEmail(email string)`**: Retrieves user by email
- **`GetUserByID(id uuid.UUID)`**: Retrieves user by ID
- **`UpdateLastSeen(userID uuid.UUID)`**: Updates user's last activity
- **`CreateMessage(senderID, receiverID uuid.UUID, content string)`**: Creates new message
- **`GetMessagesByUser(userID uuid.UUID)`**: Retrieves all messages for a user
- **`GetMessageByID(messageID uuid.UUID)`**: Retrieves message by ID
- **`GetConversation(userID1, userID2 uuid.UUID)`**: Retrieves conversation between users
- **`MarkMessageAsRead(messageID uuid.UUID)`**: Updates message read status

### postgres_test.go
Tests for PostgreSQL database operations.

- **`TestNewPostgresDB`**: Tests database connection
- **`TestCreateUser`**: Tests user creation
- **`TestGetUserByEmail`**: Tests user retrieval by email
- **`TestGetUserByID`**: Tests user retrieval by ID
- **`TestUpdateLastSeen`**: Tests last seen update
- **`TestCreateMessage`**: Tests message creation
- **`TestGetMessagesByUser`**: Tests message retrieval
- **`TestGetConversation`**: Tests conversation retrieval
- **`TestMarkMessageAsRead`**: Tests marking messages as read

### schema.sql
SQL schema definition for the database.

- **`users`**: Table for user data
  - id, username, email, password_hash, display_name, avatar_url, created_at, last_seen
- **`messages`**: Table for message data
  - id, sender_id, receiver_id, content, created_at, is_read, updated_at

## Package: internal/models

### user.go
Defines data structures and validation for user-related operations.

- **`type User struct`**: Complete user model with all fields
- **`type UserRegistration struct`**: Input validation for registration
- **`type UserLogin struct`**: Input validation for login
- **`type UserResponse struct`**: Public user data (excludes sensitive fields)

### message.go
Defines data structures for message-related operations.

- **`type Message struct`**: Complete message model with all fields
- **`type MessageRequest struct`**: Input validation for message creation
- **`type MessageResponse struct`**: Public message data with sender information

## Package: internal/websocket

### handler.go
Implements WebSocket connection handling for real-time messaging.

- **`type Client struct`**: Represents a connected WebSocket client
  - ID, Socket, Send channel
- **`type Manager struct`**: Manages active WebSocket connections
  - clients map, broadcast/register/unregister channels
- **`type WebSocketMessage struct`**: Message format for WebSocket communication
  - Type, SenderID, ReceiverID, Content, IsTyping, Token, Timestamp
- **`NewManager()`**: Creates a new WebSocket manager
- **`Run()`**: Starts the WebSocket manager event loop
  - Handles client registration/unregistration
  - Processes broadcast messages
- **`SendToUser(userID uuid.UUID, message []byte)`**: Sends message to specific user
- **`HandleWebSocket(c *gin.Context)`**: Handles WebSocket connection requests
  - Authenticates via JWT from context or WebSocket protocol
  - Upgrades HTTP connection to WebSocket
  - Registers client with manager
- **`readPump(m *Manager)`**: Reads messages from client
  - Processes different message types (auth, message, typing)
  - Routes messages to appropriate recipients
- **`writePump()`**: Writes messages to client
  - Sends queued messages
  - Handles ping/pong for connection health

### handler_test.go
Comprehensive tests for WebSocket functionality.

- **`TestNewManager`**: Tests manager creation
- **`TestManagerRun`**: Tests manager event loop
- **`TestSendToUser`**: Tests sending messages to specific users
- **`TestHandleWebSocket`**: Tests WebSocket connection handling
- **`TestWebSocketMessageExchange`**: Tests message exchange between clients
- **`TestTypingIndicator`**: Tests typing indicator functionality
- **`TestClientDisconnect`**: Tests client disconnection handling
- **`TestAuthenticationHandling`**: Tests authentication via JWT
- **`TestUnauthorizedAccess`**: Tests unauthorized access attempts
- **`TestConcurrentConnections`**: Tests multiple concurrent connections
- **`TestErrorHandling`**: Tests error handling in WebSocket operations
- **`TestPingPong`**: Tests connection health monitoring
- **`TestJWTProtocolAuthentication`**: Tests JWT authentication via WebSocket protocol
- **`TestAuthMessageHandling`**: Tests authentication message handling
- **`TestJWTHeaderAuthentication`**: Tests JWT authentication via HTTP header

## Configuration

### .env
Environment configuration file.

- **JWT_SECRET**: Secret key for signing and verifying JWT tokens
- **PORT**: Port for HTTP server (default 8080)
- **ENV**: Application environment (development, production)
- **DB_TYPE**: Database type (postgres, mysql)
- **DATABASE_URL**: Full database connection string
- **DB_HOST**: Database hostname (localhost for local development)
- **DB_PORT**: Database port (usually 5432)
- **DB_NAME**: Database name (converse)
- **DB_USER**: Database username (system username)
- **DB_PASSWORD**: Database password
- **ALLOWED_ORIGINS**: CORS allowed origins (for frontend communication)
- **LOG_LEVEL**: Application logging level
- **API_PREFIX**: Prefix for all API routes

## Authentication Flow

1. **Registration**: Client sends username, email, password → Server creates user → Returns user data
2. **Login**: Client sends email, password → Server verifies → Returns JWT token and user data
3. **Authenticated Requests**: Client includes JWT in one of three ways:
   - HTTP Authorization header (Bearer token) for REST API requests
   - WebSocket protocol field (jwt, token) for WebSocket connections
   - Authentication message over established WebSocket connection (legacy support)

## WebSocket Communication

### Connection Establishment
1. Client connects to `/api/ws` endpoint with JWT authentication
2. Server validates JWT token (from header or protocol)
3. Server upgrades connection to WebSocket
4. Server registers client in WebSocket manager

### Message Types
1. **Authentication** (`type: "auth"`): Legacy authentication method
   - Contains JWT token for validation
   - Used as fallback for clients that can't set headers
2. **Message** (`type: "message"`): Regular chat message
   - Contains sender ID, receiver ID, content, timestamp
   - Stored in database and forwarded to recipient
3. **Typing** (`type: "typing"`): Typing indicator
   - Contains sender ID, receiver ID, isTyping flag
   - Forwarded to recipient without database storage
4. **Read Receipt** (`type: "read_receipt"`): Message read notification
   - Contains message ID, sender ID, receiver ID
   - Sent when recipient marks message as read

### Real-time Features
1. **Instant Messaging**: Messages delivered immediately to online users
2. **Typing Indicators**: Real-time notification when user is typing
3. **Online Presence**: Connection status tracking
4. **Read Receipts**: Notification when messages are read

## Testing Status

### Test Results Summary
1. **API Package Tests** (`internal/api/`)
   - ✅ `TestRegister`: All cases pass
   - ✅ `TestLogin`: All cases pass
   - ✅ `TestAuthMiddleware`: All cases pass
   - ✅ `TestGetMe`: All cases pass
   - ✅ `TestCreateMessage`: All cases pass
   - ✅ `TestGetMessages`: All cases pass
   - ✅ `TestGetConversation`: All cases pass
   - ✅ `TestMarkMessageAsRead`: All cases pass
   - ✅ `TestWebSocketMessageFormat`: All cases pass

2. **Database Package Tests** (`internal/database/`)
   - ✅ `TestNewPostgresDB`: Connection handling
   - ✅ `TestCreateUser`: User creation and validation
   - ✅ `TestGetUserByEmail`: User retrieval by email
   - ✅ `TestGetUserByID`: User retrieval by ID
   - ✅ `TestUpdateLastSeen`: Last activity tracking
   - ✅ `TestCreateMessage`: Message creation
   - ✅ `TestGetMessagesByUser`: Message retrieval
   - ✅ `TestGetConversation`: Conversation retrieval
   - ✅ `TestMarkMessageAsRead`: Message read status update

3. **Auth Package Tests** (`internal/auth/`)
   - ✅ `TestPasswordHashing`: Password security
   - ✅ `TestCheckPasswordHash`: Hash verification
   - ✅ `TestGenerateToken`: JWT generation
   - ✅ `TestValidateToken`: JWT validation
   - ✅ `TestInitJWTKey`: JWT key management

4. **WebSocket Package Tests** (`internal/websocket/`)
   - ✅ `TestNewManager`: Manager creation
   - ✅ `TestManagerRun`: Event loop functionality
   - ✅ `TestSendToUser`: Targeted message delivery
   - ✅ `TestHandleWebSocket`: Connection handling
   - ✅ `TestWebSocketMessageExchange`: Message exchange
   - ✅ `TestTypingIndicator`: Typing notifications
   - ✅ `TestClientDisconnect`: Disconnection handling
   - ✅ `TestAuthenticationHandling`: JWT authentication
   - ✅ `TestJWTProtocolAuthentication`: Protocol-based auth
   - ✅ `TestJWTHeaderAuthentication`: Header-based auth

5. **End-to-End API Testing**
   - ✅ User registration endpoint successfully creates new users
   - ✅ Login endpoint verifies credentials and returns valid JWT tokens
   - ✅ Protected `/me` endpoint correctly returns user data when authenticated
   - ✅ Authentication middleware properly blocks unauthorized requests
   - ✅ WebSocket connections authenticate properly with JWT
   - ✅ Real-time message delivery works between connected clients
   - ✅ Health check endpoint confirms server status

### Environment Requirements
- PostgreSQL database with user-specific credentials
- Environment variables properly configured
- JWT secret key initialized
- Database schema with users and messages tables created

### Known Issues
1. Database configuration:
   - Using system username instead of 'postgres'
   - Proper database permissions
   - Test database (`converse_test`) created

## UI Implementation

### Current Implementation Status

The frontend UI for Converse has been implemented with a focus on creating a modern, responsive chat interface. The implementation follows a component-based architecture using React, with a clean separation of concerns between different UI elements.

#### Project Structure
```
converse-ui/
├── public/
│   ├── index.html
│   └── assets/
├── src/
│   ├── components/
│   │   ├── auth/                # Authentication components
│   │   │   ├── LoginForm.jsx    # User login form
│   │   │   └── RegisterForm.jsx # User registration form
│   │   ├── layout/              # Layout components
│   │   │   └── Navbar.jsx       # Navigation bar with authentication state
│   │   └── messaging/           # Chat interface components
│   │       ├── Messaging.jsx    # Main messaging container
│   │       ├── ConversationsList.jsx # List of chat conversations
│   │       ├── ChatWindow.jsx   # Message display area
│   │       └── MessageInput.jsx # Message input field
│   ├── pages/                   # Page components
│   │   ├── LoginPage.jsx        # Login page
│   │   ├── RegisterPage.jsx     # Registration page
│   │   └── ChatPage.jsx         # Main chat page
│   ├── context/                 # React context providers
│   │   └── AuthContext.jsx      # Authentication state management
│   ├── services/                # API services
│   │   ├── api.js               # REST API client
│   │   └── websocket.js         # WebSocket client with JWT authentication
│   ├── utils/                   # Utility functions
│   ├── routes/                  # Route definitions
│   ├── assets/                  # Static assets
│   ├── index.css                # Global styles
│   ├── App.jsx                  # Main application component
│   └── main.jsx                 # Application entry point
└── package.json                 # Dependencies and scripts
```

#### Key Features Implemented

1. **Authentication System**
   - User registration with form validation
   - User login with credential verification
   - JWT-based authentication
   - Protected routes for authenticated users
   - Logout functionality with confirmation dialog

2. **Navigation and Layout**
   - Responsive navigation bar
   - Conditional rendering based on authentication state
   - User welcome message with username display
   - Logout confirmation dialog to prevent accidental logouts

3. **Messaging Interface**
   - Conversation list with contact avatars and preview messages
   - Chat window with message bubbles and timestamps
   - Message input with send functionality
   - Visual distinction between sent and received messages
   - Contact header showing name, status, and avatar
   - Typing indicators for real-time feedback
   - Read receipts showing when messages are seen

4. **WebSocket Integration**
   - Secure WebSocket connection with JWT authentication
   - Real-time message delivery
   - Typing indicators
   - Connection status management
   - Automatic reconnection on disconnection
   - Read receipt notifications

5. **Styling and User Experience**
   - Modern, clean UI with consistent color scheme
   - Responsive design that works on mobile and desktop
   - Animations for message appearance and typing indicators
   - Visual feedback for user interactions
   - Consistent styling across all components

#### WebSocket Client Implementation

The WebSocket client implementation includes:

1. **Connection Establishment**
   ```javascript
   // Connect with JWT authentication
   const connectWebSocket = (token) => {
     // Close existing connection if any
     if (socket && socket.readyState !== WebSocket.CLOSED) {
       socket.close();
     }
     
     // Create new connection with JWT in protocol
     socket = new WebSocket(`ws://localhost:8080/api/ws`, ['jwt', token]);
     
     // Set up event handlers
     socket.onopen = handleOpen;
     socket.onmessage = handleMessage;
     socket.onclose = handleClose;
     socket.onerror = handleError;
   };
   ```

2. **Message Handling**
   ```javascript
   // Handle incoming WebSocket messages
   const handleMessage = (event) => {
     const message = JSON.parse(event.data);
     
     switch (message.type) {
       case 'message':
         // Handle new chat message
         addMessageToConversation(message);
         break;
       case 'typing':
         // Handle typing indicator
         updateTypingStatus(message);
         break;
       case 'read_receipt':
         // Handle read receipt
         updateMessageReadStatus(message);
         break;
     }
   };
   ```

3. **Sending Messages**
   ```javascript
   // Send message via WebSocket
   const sendMessage = (receiverId, content) => {
     if (socket && socket.readyState === WebSocket.OPEN) {
       const message = {
         type: 'message',
         receiver_id: receiverId,
         content: content,
         timestamp: new Date()
       };
       
       socket.send(JSON.stringify(message));
     }
   };
   ```

4. **Typing Indicators**
   ```javascript
   // Send typing indicator
   const sendTypingIndicator = (receiverId, isTyping) => {
     if (socket && socket.readyState === WebSocket.OPEN) {
       const message = {
         type: 'typing',
         receiver_id: receiverId,
         is_typing: isTyping,
         timestamp: new Date()
       };
       
       socket.send(JSON.stringify(message));
     }
   };
   ```

### Future UI Enhancements

1. **Enhanced Real-time Features**
   - User presence indicators (online/offline/away)
   - Message delivery status (sent/delivered/read)
   - Group chat functionality
   - Voice and video calling

2. **Enhanced User Experience**
   - Message search functionality
   - File and image sharing
   - User profile customization
   - Theme selection (light/dark mode)
   - Notification system

3. **Performance Optimizations**
   - Virtualized lists for better performance with large message histories
   - Lazy loading of images and media
   - Optimized rendering for mobile devices
   - Offline support with message queuing 