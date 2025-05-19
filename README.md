# System Architecture
Converse follows a client-server architecture with real-time communication capabilities through WebSockets. The system consists of two primary components:

*   **React Frontend**: Handles user interface, state management, and client-side WebSocket connections
*   **Go Backend**: Provides API endpoints, handles authentication, processes messages, and manages WebSocket connections

```mermaid
graph TD
    subgraph Client Browser
        A[Converse UI Frontend]
        B[WebSocket Client]
    end

    subgraph Server
        C[Go Backend API]
        D[WebSocket Server]
        E[Database]
    end

    A --- C
    A --> B
    B --- D
    C --- E
    D --- E
```