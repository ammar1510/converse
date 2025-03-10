# Codebase Structure

## Database Package

The database package provides an abstraction layer for database operations. It defines interfaces and implementations for various database systems.

### DBInterface

`DBInterface` defines the contract that all database implementations must fulfill. It provides methods for user and message operations.

#### User Methods

- **CreateUser(username, email, passwordHash string) (*models.User, error)**
  - Creates a new user in the database
  - Parameters: username, email, and hashed password
  - Returns: the created user object or an error

- **GetUserByEmail(email string) (*models.User, error)**
  - Retrieves a user by their email address
  - Parameters: email address to search for
  - Returns: user object if found, or an error

- **GetUserByID(id uuid.UUID) (*models.User, error)**
  - Retrieves a user by their UUID
  - Parameters: user ID to search for
  - Returns: user object if found, or an error

- **UpdateLastSeen(userID uuid.UUID) error**
  - Updates the last seen timestamp for a user
  - Parameters: ID of the user to update
  - Returns: error if the operation fails

#### Message Methods

- **CreateMessage(senderID, receiverID uuid.UUID, content string) (*models.Message, error)**
  - Creates a new message between users
  - Parameters: sender's ID, receiver's ID, and message content
  - Returns: the created message object or an error

- **GetMessagesByUser(userID uuid.UUID) ([]*models.Message, error)**
  - Retrieves all messages for a specific user
  - Parameters: ID of the user whose messages to retrieve
  - Returns: array of message objects or an error

- **GetMessageByID(messageID uuid.UUID) (*models.Message, error)**
  - Retrieves a specific message by its ID
  - Parameters: ID of the message to retrieve
  - Returns: message object if found, or an error

- **GetConversation(userID1, userID2 uuid.UUID) ([]*models.Message, error)**
  - Retrieves all messages between two users (a conversation)
  - Parameters: IDs of the two users in the conversation
  - Returns: array of message objects representing the conversation, or an error

- **MarkMessageAsRead(messageID uuid.UUID) error**
  - Marks a message as read
  - Parameters: ID of the message to mark
  - Returns: error if the operation fails

#### Direct Database Access

- **Exec(query string, args ...interface{}) (ExecResult, error)**
  - Executes a raw SQL query
  - Parameters: SQL query string and optional arguments
  - Returns: execution result interface or an error

#### Common Methods

- **Close() error**
  - Closes the database connection
  - Returns: error if the close operation fails

### Supporting Types and Functions

- **ExecResult Interface**: Defines methods for SQL execution results
- **DatabaseType**: Enum representing supported database types (PostgreSQL, MySQL)
- **NewDatabase**: Factory function that returns the appropriate database implementation based on the type
