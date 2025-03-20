# Converse WebSocket API Documentation

This document provides comprehensive guidance on how to integrate with the Converse WebSocket API for real-time messaging.

## WebSocket Endpoints

### Primary Endpoint (Authenticated)

- **URL**: `/api/ws`
- **Authentication**: Required (JWT Bearer Token or URL parameter)
- **Purpose**: Main endpoint for authenticated real-time messaging
- **Usage**: Production use for authenticated users

### Test/Alternative Endpoint

- **URL**: `/socket`
- **Authentication**: Optional (token as URL parameter)
- **Purpose**: Supports both authenticated and unauthenticated connections
- **Usage**: Testing, debugging, or fallback connections

## Authentication

### Method 1: HTTP Header Authentication (Primary)

For the `/api/ws` endpoint, authentication can be performed using a JWT Bearer token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Method 2: URL Parameter Authentication

Both the `/api/ws` and `/socket` endpoints support authentication using a token URL parameter:

```
ws://localhost:8080/api/ws?token=<your-jwt-token>
```

```
ws://localhost:8080/socket?token=<your-jwt-token>
```

### Obtaining a Token

To obtain a JWT token, make a POST request to the login endpoint:

```
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

The response will include a token that can be used for WebSocket authentication:

```json
{
  "token": "your-jwt-token",
  "expiry": "2023-03-21T10:00:00Z",
  "user": {
    "id": "user-uuid",
    "username": "username",
    ...
  }
}
```

## Message Format

### Sending Messages

When sending messages through the WebSocket, use the following JSON format:

```json
{
  "type": "message",
  "receiver_id": "recipient-uuid",
  "content": "Your message text"
}
```

The server will automatically add the `sender_id` and `timestamp` fields.

### Sending Typing Indicators

```json
{
  "type": "typing",
  "receiver_id": "recipient-uuid",
  "is_typing": true
}
```

Set `is_typing` to `false` when the user stops typing.

### Receiving Messages

Messages received from the server will have this format:

```json
{
  "type": "message",
  "sender_id": "sender-uuid",
  "receiver_id": "recipient-uuid",
  "content": "Message text",
  "timestamp": "2023-03-20T10:04:21.709455+05:30"
}
```

### Error Messages

Error messages from the server follow this format:

```json
{
  "type": "error",
  "content": "Error description",
  "timestamp": "2023-03-20T10:04:21.709455+05:30"
}
```

Common error messages include:
- "Invalid message format"
- "Invalid receiver ID"
- "Unknown message type"

## Connection Examples

### Command Line with wscat/websocat

#### Authenticated Connection to Primary Endpoint (Using Header)

```bash
# Step 1: Get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your_email@example.com","password":"your_password"}' \
  | jq -r '.token')

# Step 2: Connect with the token in Authorization header
websocat "ws://localhost:8080/api/ws" -H "Authorization: Bearer $TOKEN"
```

#### Authenticated Connection to Primary Endpoint (Using URL Parameter)

```bash
# Step 1: Get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your_email@example.com","password":"your_password"}' \
  | jq -r '.token')

# Step 2: Connect with the token in URL parameter
websocat "ws://localhost:8080/api/ws?token=$TOKEN"
```

#### Connection to Alternative Endpoint

```bash
# Using token in URL parameter
websocat "ws://localhost:8080/socket?token=$TOKEN"
```

### JavaScript/Browser Client Integration

```javascript
// Get token from login request (pseudocode)
const token = await loginAndGetToken();

// Method 1: Connect to WebSocket with token in URL parameter (recommended for browser clients)
const socket = new WebSocket(`ws://localhost:8080/api/ws?token=${token}`);

socket.onopen = () => {
  console.log('WebSocket connection established');
  
  // Send a message
  const message = {
    type: 'message',
    receiver_id: 'recipient-uuid',
    content: 'Hello!'
  };
  socket.send(JSON.stringify(message));
};

// Method 2: Connect to WebSocket with Authorization header (if supported by your browser/environment)
// Note: Some browsers may not support setting headers for WebSocket connections
const socketWithHeader = new WebSocket('ws://localhost:8080/api/ws');

// Attempt to set headers (may not work in all environments)
// This is implementation-specific and might require a library
if (socketWithHeader.setRequestHeader) {
  socketWithHeader.setRequestHeader('Authorization', `Bearer ${token}`);
}

// Handle incoming messages
socket.onmessage = (event) => {
  try {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
    
    switch (message.type) {
      case 'message':
        // Handle chat message
        displayMessage(message);
        break;
      case 'typing':
        // Handle typing indicator
        updateTypingStatus(message);
        break;
      case 'error':
        // Handle error
        console.error('WebSocket error:', message.content);
        break;
      default:
        console.warn('Unknown message type:', message.type);
    }
  } catch (error) {
    console.error('Error parsing message:', error);
  }
};

// Handle connection errors
socket.onerror = (error) => {
  console.error('WebSocket error:', error);
};

// Handle connection close
socket.onclose = (event) => {
  console.log('WebSocket connection closed', event.code, event.reason);
  // Implement reconnection logic here
};
```

## Important Implementation Notes

1. **Message Validation**:
   - All messages must be valid JSON
   - The `type` field is required
   - The `receiver_id` must be a valid UUID

2. **Rate Limiting**:
   - The server limits messages to 60 per minute per client

3. **Ping/Pong Protocol**:
   - The server sends ping frames every 54 seconds
   - Clients must respond with pong frames
   - Read deadline is extended by 60 seconds after each pong

4. **Connection Lifecycle**:
   - Connections may be closed if:
     - Client doesn't respond to pings
     - Rate limits are exceeded
     - Invalid messages are repeatedly sent

5. **Error Handling**:
   - Always handle error messages from the server
   - Implement reconnection logic in your client

## HTTP Endpoints for Messages

In addition to the WebSocket API, the following HTTP endpoints are available for message management:

- `POST /api/messages` - Send a new message
- `GET /api/messages` - Get all messages for the authenticated user
- `GET /api/messages/conversation/:userID` - Get conversation with a specific user
- `PUT /api/messages/:messageID/read` - Mark a message as read

These HTTP endpoints use the same JWT authentication mechanism as the WebSocket API. 