# Converse - Real-Time Chat Application

A high-performance real-time chat service built with Go and React, designed for direct messaging between users.

## Project Overview

This project is a scalable WebSocket-based chat application with features similar to modern messaging platforms. It demonstrates proficiency in Go backend development, React frontend development, handling concurrent connections, real-time data processing, and state management.

## Architecture

```
┌─────────────┐     ┌───────────────────────┐      ┌────────────────┐
│ React       │◄────┤ Go Chat Service       │◄─────┤PostgreSQL      │
│ Frontend    │     │ (WebSockets + REST API)│      │(User/Message DB)│
└─────────────┘     └───────────────────────┘      └────────────────┘
```

### Core Components

1. **Go Backend**: 
   - REST API for authentication and data retrieval
   - WebSocket server for real-time communication
   - JWT-based authentication

2. **React Frontend**:
   - Modern UI with components for chats and messages
   - Real-time updates via WebSocket
   - Context API for state management

3. **PostgreSQL Database**:
   - User management
   - Message persistence
   - Conversation tracking

## Authentication Flow

```
1. Registration: User → POST /api/auth/register → Store in PostgreSQL
2. Login: User → POST /api/auth/login → Verify credentials → Return JWT
3. JWT Storage: Client stores token in localStorage
4. Authenticated Requests: Include JWT in Authorization header
5. WebSocket Auth: Establish connection with JWT via URL parameter
```

## Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Direct messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    sender_id UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## Features

### Core Features

- **User Management**
  - Registration and authentication
  - User profiles
  - Online/offline status

- **Real-Time Messaging**
  - Direct messaging between users
  - Read receipts
  - Message history

## API Endpoints

### User Management
- `POST /api/auth/register` - Create a new user account
- `POST /api/auth/login` - Authenticate and receive JWT token
- `GET /api/auth/me` - Get current user profile
- `GET /api/users` - List users

### Message Management
- `GET /api/messages` - Get all messages for the current user
- `POST /api/messages` - Send a new message
- `GET /api/messages/conversation/:userID` - Get conversation with specific user
- `PUT /api/messages/:messageID/read` - Mark a message as read

### WebSocket Interface
- `WS /api/ws` - Main WebSocket endpoint (authenticated via token in URL)

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Framework**: Gin
- **Database**: PostgreSQL
- **WebSockets**: Gorilla WebSocket
- **Authentication**: JWT tokens

### Frontend
- **Framework**: React
- **State Management**: Context API
- **Styling**: CSS/TailwindCSS
- **HTTP Client**: Axios
- **WebSockets**: Native WebSocket API

## Project Structure

### Backend
```
converse/
├── cmd/
│   └── server/                # Entry point
│       └── main.go
├── internal/
│   ├── api/                   # REST API handlers
│   │   ├── auth.go            # Auth endpoints
│   │   ├── messages.go        # Message endpoints
│   │   └── router.go          # Route configuration
│   ├── auth/                  # Auth logic
│   ├── database/              # DB interactions
│   │   └── postgres.go
│   ├── models/                # Data structures
│   └── websocket/             # WebSocket handling
├── pkg/                       # Shared utilities
└── go.mod
```

### Frontend
```
converse-ui/
├── src/
│   ├── components/            # Reusable UI components
│   ├── context/               # React Context providers
│   ├── pages/                 # Page components
│   ├── services/              # API service modules
│   │   ├── api.js             # Axios instance
│   │   ├── authService.js     # Authentication
│   │   ├── messageService.js  # Message handling
│   │   └── websocketService.js # WebSocket client
│   ├── utils/                 # Utility functions
│   └── App.jsx                # Main application component
├── public/
└── package.json
```

## WebSocket Protocol

WebSocket connections are established at `/api/ws` with authentication via JWT token in the URL parameter.

Message format example:
```json
{
  "type": "message",
  "sender_id": "user-uuid",
  "receiver_id": "recipient-uuid",
  "content": "Hello there!",
  "timestamp": "2023-03-20T12:34:56Z"
}
```

## Getting Started

### Backend Setup
```bash
# Run the server
go run cmd/server/main.go
```

### Frontend Setup
```bash
cd converse-ui
npm install
npm run dev
```

## License

MIT 