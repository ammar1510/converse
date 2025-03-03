# Go Chat Service

A high-performance real-time chat service built with Go, designed for small-scale deployment and demonstration purposes.

## Project Overview

This project is a scalable WebSocket-based chat application with features similar to Slack or Discord. It demonstrates proficiency in Go development, handling concurrent connections, real-time data processing, and state management.

## Architecture

```
┌─────────────┐     ┌───────────────────────┐      ┌────────────────┐
│ Web/Mobile  │◄────┤ Go Chat Service       │◄─────┤Supabase Postgres│
│ Client      │     │ (WebSockets + REST API)│      │(User/Message DB)│
└─────────────┘     └───────────┬───────────┘      └────────────────┘
                               │
                               ▼
                      ┌────────────────┐
                      │ Upstash Redis  │
                      │(Presence/Cache)│
                      └────────────────┘
```

### Core Components

1. **Go Server**: Single monolithic server handling both REST API and WebSockets
2. **WebSocket Handler**: Manages real-time bidirectional communication
3. **Message Broker**: Routes messages between users and channels
4. **Authentication**: JWT-based auth system
5. **Supabase PostgreSQL**: For user, channel, and message persistence
6. **Upstash Redis**: For user presence and temporary data

## Authentication Flow

```
1. Registration: User → POST /api/auth/register → Store in Supabase
2. Login: User → POST /api/auth/login → Verify credentials → Return JWT
3. JWT Storage: Client stores token in localStorage/sessionStorage
4. Authenticated Requests: Include JWT in Authorization header
5. WebSocket Auth: Establish connection with JWT for real-time features
```

## Request Processing Flow

```
REST API Request Flow:
User → API Endpoint → Handler → Service Logic → Database → Response

WebSocket Flow:
User → Connect WebSocket → Auth → Message Handler → Broker → Recipients
```

## Database Schema

```sql
-- Core tables for MVP
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE channels (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  is_private BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_by UUID REFERENCES users(id)
);

CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id),
  content TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE channel_members (
  channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (channel_id, user_id)
);
```

## Features

### Core Features (MVP)

- **User Management**
  - Registration and authentication
  - User profiles
  - Online/offline status

- **Channel Operations**
  - Create/join/leave channels
  - Public channels

- **Real-Time Messaging**
  - Text messages with timestamps
  - Message history

### Advanced Features (If Time Permits)

- Private channels and direct messaging
- Typing indicators
- Message delivery confirmations
- Message reactions
- Simple file sharing
- User presence (away, etc.)
- Read receipts

## WebSocket Usage

WebSockets are used for all real-time communications:
- Receiving/sending messages
- User presence updates
- Channel join/leave notifications
- Typing indicators (if implemented)
- Read receipts (if implemented)

REST API is used for:
- Authentication
- User profile management
- Channel management
- Fetching message history

## API Endpoints

### User Management
- `POST /api/auth/register` - Create a new user account
- `POST /api/auth/login` - Authenticate and receive JWT token
- `GET /api/auth/me` - Get current user profile
- `GET /api/users` - List users (with pagination)

### Channel Management
- `GET /api/channels` - List available channels
- `POST /api/channels` - Create a new channel
- `GET /api/channels/:id` - Get channel details
- `POST /api/channels/:id/members` - Join channel

### Message Management
- `GET /api/channels/:id/messages` - Fetch message history
- `POST /api/channels/:id/messages` - Post a new message (fallback)

### WebSocket Interface
- `WS /ws` - Main WebSocket endpoint (authenticated)

## Technology Stack

- **Backend**: Go (Golang)
- **Database**: 
  - Supabase PostgreSQL (free tier)
  - Upstash Redis (free tier)
- **API Framework**: Gin or Echo
- **WebSockets**: Gorilla WebSocket package
- **Authentication**: JWT tokens
- **Frontend** (optional): Simple React/Vue app

## Resource Usage (Estimates for 20 Users)

- **Database storage**: ~10-20MB (well within 500MB free tier)
- **Redis commands**: ~500-1,000/day (well within 10,000/day free tier)
- **Bandwidth**: ~100-200MB/month (well within 2GB free tier)

## Project Structure

```
my-chat-service/
├── cmd/
│   └── server/                # Entry point
│       └── main.go
├── internal/
│   ├── api/                   # REST API handlers
│   │   ├── auth.go            # Auth endpoints
│   │   ├── channels.go        # Channel endpoints
│   │   ├── messages.go        # Message endpoints
│   │   └── router.go          # Route configuration
│   ├── auth/                  # Auth logic
│   │   ├── jwt.go
│   │   └── password.go
│   ├── config/
│   ├── database/              # DB interactions
│   │   ├── supabase.go
│   │   └── redis.go
│   ├── models/                # Data structures
│   │   ├── user.go
│   │   ├── channel.go
│   │   └── message.go
│   ├── websocket/             # WebSocket handling
│   │   ├── client.go
│   │   ├── hub.go
│   │   └── message.go
│   └── broker/                # Message routing
├── pkg/                       # Shared utilities
├── web/                       # Frontend (optional)
├── go.mod
├── go.sum
└── docker-compose.yml
```

## Implementation Plan

### Phase 1: Core Setup
- Project scaffolding
- Database schema design in Supabase
- Basic user authentication
- Connection to Upstash Redis

### Phase 2: Basic Messaging
- WebSocket connection handling
- User-to-user messaging
- Channel creation and basic management
- Message persistence in Supabase

### Phase 3: Enhanced Features
- Implement remaining REST API endpoints
- User presence with Redis
- Message history and pagination

### Phase 4: Optional Features
- Simple web client
- Additional features as time permits

## Setup Instructions

### Prerequisites
- Go 1.20+
- Supabase account (free tier)
- Upstash account (free tier)
- Docker and Docker Compose (for local development)

### Development Setup
```bash
# Clone the repository
git clone https://github.com/yourusername/go-chat-service.git
cd go-chat-service

# Install dependencies
go mod download

# Set up local environment
cp .env.example .env
# Edit .env with your Supabase and Upstash credentials

# Run locally
go run cmd/server/main.go
```

## WebSocket Protocol

WebSocket connections will be established at `/ws` with authentication via JWT token in the request header or query parameter.

Message format:
```json
{
  "type": "message|typing|join|leave",
  "channel": "channel_id",
  "content": "message content",
  "timestamp": 1634000000
}
```

## License

MIT 