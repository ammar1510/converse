import { WS_URL } from '../config';
import { getToken } from '../utils/tokenStorage';

// Connection states for state machine
const CONNECTION_STATES = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting'
};

// Standard WebSocket event types
const EVENT_TYPES = {
  // WebSocket connection events
  CONNECT: 'connect',
  DISCONNECT: 'disconnect',
  RECONNECTING: 'reconnecting',
  RECONNECT_FAILED: 'reconnect_failed',
  ERROR: 'error',
  
  // Application-specific events
  MESSAGE: 'message',
  TYPING: 'typing'
};

/**
 * Service for managing WebSocket connections and real-time messaging
 */
class WebSocketService {
  constructor() {
    this.socket = null;
    this.eventRegistry = {}; // Registry of event handlers by type
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectTimeout = null;
    this.currentToken = null;
    
    // State machine for connection status
    this.connectionState = CONNECTION_STATES.DISCONNECTED;
    
    // Use the URL from config
    this.wsUrl = WS_URL;
    
    // Initialize event registry with empty arrays for all event types
    Object.values(EVENT_TYPES).forEach(eventType => {
      this.eventRegistry[eventType] = [];
    });
  }
  
  /**
   * Connect to the WebSocket server
   * @param {string} token - JWT authentication token
   * @param {boolean} isReconnect - Whether this is a reconnection attempt
   */
  connect(token, isReconnect = false) {
    // Validate token
    if (!token) {
      console.error('Cannot connect to WebSocket: No token provided');
      this.triggerEvent(EVENT_TYPES.ERROR, { message: 'No token provided' });
      return;
    }
    
    // Check current state and token
    if (this.connectionState === CONNECTION_STATES.CONNECTED || 
        this.connectionState === CONNECTION_STATES.CONNECTING) {
      // If the token has changed, disconnect and reconnect with the new token
      if (token !== this.currentToken) {
        console.log('Token changed, reconnecting with new token');
        this.disconnect();
      } else {
        console.log(`Already ${this.connectionState}, ignoring connect request`);
        return;
      }
    }
    
    // Update state and token
    this.connectionState = isReconnect ? 
      CONNECTION_STATES.RECONNECTING : 
      CONNECTION_STATES.CONNECTING;
    this.currentToken = token;
    
    // Attempt to establish connection
    try {
      console.log(`Connecting to WebSocket at ${this.wsUrl} (${this.connectionState})`);
      
      // Create WebSocket URL with token
      const wsUrlWithToken = `${this.wsUrl}?token=${encodeURIComponent(token)}`;
      this.socket = new WebSocket(wsUrlWithToken);
      
      // Set up event handlers
      this.setupSocketEventHandlers();
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.connectionState = CONNECTION_STATES.DISCONNECTED;
      this.handleConnectionFailure(token);
    }
  }
  
  /**
   * Set up the event handlers for the WebSocket
   */
  setupSocketEventHandlers() {
    if (!this.socket) return;
    
    // Connection successfully established
    this.socket.onopen = () => {
      this.connectionState = CONNECTION_STATES.CONNECTED;
      this.reconnectAttempts = 0;
      console.log('WebSocket connection established successfully');
      
      // Trigger connect event
      this.triggerEvent(EVENT_TYPES.CONNECT, {});
    };
    
    // Message received
    this.socket.onmessage = (event) => {
      console.log('WebSocket message received');
      try {
        const data = JSON.parse(event.data);
        
        // Handle different message types
        if (data.type === 'typing') {
          this.triggerEvent(EVENT_TYPES.TYPING, data);
        } else {
          // Default to 'message' type
          this.triggerEvent(EVENT_TYPES.MESSAGE, data);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        this.triggerEvent(EVENT_TYPES.ERROR, { 
          message: 'Failed to parse message', 
          error,
          originalData: event.data 
        });
      }
    };
    
    // Connection closed
    this.socket.onclose = (event) => {
      console.log(`WebSocket disconnected: ${event.code} ${event.reason}`);
      const wasConnected = this.connectionState === CONNECTION_STATES.CONNECTED;
      this.socket = null;
      
      // Update state
      this.connectionState = CONNECTION_STATES.DISCONNECTED;
      
      // Trigger disconnect event
      this.triggerEvent(EVENT_TYPES.DISCONNECT, { 
        code: event.code, 
        reason: event.reason 
      });
      
      // Handle reconnection based on close code
      this.handleCloseEvent(event, wasConnected);
    };
    
    // Error occurred
    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.triggerEvent(EVENT_TYPES.ERROR, { error });
    };
  }
  
  /**
   * Handle WebSocket close event and determine reconnection strategy
   * @param {CloseEvent} event - WebSocket close event
   * @param {boolean} wasConnected - Whether the socket was previously in connected state
   */
  handleCloseEvent(event, wasConnected) {
    // Don't attempt to reconnect if:
    // 1. Clean disconnect (code 1000)
    // 2. Max reconnect attempts reached
    if (event.code === 1000) {
      console.log('Clean disconnect, not attempting to reconnect');
      return;
    }
    
    // For other disconnects, try to refresh and reconnect
    console.log(`Abnormal close (${event.code}), handling reconnection`);
    
    // Only try to reconnect if:
    // 1. We were previously connected, or
    // 2. We haven't reached max reconnect attempts
    if (wasConnected || this.reconnectAttempts < this.maxReconnectAttempts) {
      this.handleConnectionFailure(this.currentToken);
    } else {
      console.log('Max reconnect attempts reached, giving up');
      this.triggerEvent(EVENT_TYPES.RECONNECT_FAILED, { 
        attempts: this.reconnectAttempts 
      });
    }
  }
  
  /**
   * Handle connection failure with exponential backoff
   * @param {string} token - JWT authentication token
   */
  handleConnectionFailure(token) {
    // Don't attempt to reconnect if we've reached the max attempts
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached');
      this.triggerEvent(EVENT_TYPES.RECONNECT_FAILED, { 
        attempts: this.reconnectAttempts 
      });
      return;
    }
    
    // First try to get a fresh token
    const freshToken = getToken();
    const tokenToUse = freshToken || token;
    
    if (!tokenToUse) {
      console.log('No token available for reconnection - user may be logged out');
      return;
    }
    
    // Exponential backoff with jitter
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    const jitter = Math.random() * 1000;
    this.reconnectAttempts++;
    
    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${Math.floor((delay + jitter) / 1000)}s...`);
    
    // Clear any existing timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    
    // Set up new timeout for reconnection
    this.reconnectTimeout = setTimeout(() => {
      console.log('Reconnecting with token:', tokenToUse ? tokenToUse.substring(0, 10) + '...' : 'none');
      this.connect(tokenToUse, true);
    }, delay + jitter);
    
    this.triggerEvent(EVENT_TYPES.RECONNECTING, { 
      attempt: this.reconnectAttempts, 
      maxAttempts: this.maxReconnectAttempts 
    });
  }
  
  /**
   * Disconnect from the WebSocket server
   */
  disconnect() {
    if (this.socket) {
      this.socket.close(1000, 'User disconnected');
      this.socket = null;
    }
    
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    this.connectionState = CONNECTION_STATES.DISCONNECTED;
    this.currentToken = null;
    this.reconnectAttempts = 0;
  }
  
  /**
   * Register a handler for a specific event type
   * @param {string} eventType - Event type from EVENT_TYPES
   * @param {function} handler - Event handler function
   * @returns {function} Unsubscribe function to remove this handler
   */
  on(eventType, handler) {
    // Validate event type
    if (!Object.values(EVENT_TYPES).includes(eventType)) {
      console.warn(`Unknown event type: ${eventType}`);
      return () => {}; // Return no-op unsubscribe function
    }
    
    // Add handler to registry
    this.eventRegistry[eventType].push(handler);
    
    // Return unsubscribe function
    return () => this.off(eventType, handler);
  }
  
  /**
   * Remove a handler for a specific event type
   * @param {string} eventType - Event type from EVENT_TYPES
   * @param {function} handler - Event handler function to remove
   */
  off(eventType, handler) {
    if (this.eventRegistry[eventType]) {
      this.eventRegistry[eventType] = this.eventRegistry[eventType].filter(h => h !== handler);
    }
  }
  
  /**
   * Trigger an event to all registered handlers
   * @param {string} eventType - Event type from EVENT_TYPES
   * @param {object} data - Event data
   */
  triggerEvent(eventType, data) {
    if (!this.eventRegistry[eventType] || this.eventRegistry[eventType].length === 0) {
      console.log(`No handlers registered for ${eventType} event`);
      return;
    }
    
    console.log(`Triggering ${eventType} event to ${this.eventRegistry[eventType].length} handlers`);
    
    // Copy handlers array to prevent modification during iteration
    const handlers = [...this.eventRegistry[eventType]];
    
    // Call each handler with data
    handlers.forEach(handler => {
      try {
        handler(data);
      } catch (error) {
        console.error(`Error in ${eventType} event handler:`, error);
      }
    });
  }
  
  /**
   * Legacy addEventListener for backward compatibility
   * @param {string} event - Event name
   * @param {function} callback - Event callback
   * @deprecated Use on() instead
   */
  addEventListener(event, callback) {
    return this.on(event, callback);
  }
  
  /**
   * Legacy removeEventListener for backward compatibility
   * @param {string} event - Event name
   * @param {function} callback - Event callback to remove
   * @deprecated Use off() instead
   */
  removeEventListener(event, callback) {
    this.off(event, callback);
  }
  
  /**
   * Legacy dispatchEvent for backward compatibility
   * @param {string} event - Event name
   * @param {object} data - Event data
   * @deprecated Use triggerEvent() instead
   */
  dispatchEvent(event, data) {
    this.triggerEvent(event, data);
  }
  
  /**
   * Send a message via WebSocket
   * @param {string} receiverId - UUID of the message recipient
   * @param {string} content - Message content
   * @returns {boolean} Success status
   */
  sendMessage(receiverId, content) {
    if (!this.isConnected()) {
      console.error('WebSocket not connected, cannot send message');
      return false;
    }
    
    try {
      this.socket.send(JSON.stringify({
        type: 'message',
        receiver_id: receiverId,
        content: content
      }));
      return true;
    } catch (error) {
      console.error('Error sending message:', error);
      this.triggerEvent(EVENT_TYPES.ERROR, {
        message: 'Failed to send message',
        error
      });
      return false;
    }
  }
  
  /**
   * Send typing indicator via WebSocket
   * @param {string} receiverId - UUID of the message recipient
   * @param {boolean} isTyping - Whether the user is typing
   * @returns {boolean} Success status
   */
  sendTyping(receiverId, isTyping) {
    if (!this.isConnected()) {
      return false;
    }
    
    try {
      this.socket.send(JSON.stringify({
        type: 'typing',
        receiver_id: receiverId,
        is_typing: isTyping
      }));
      return true;
    } catch (error) {
      console.error('Error sending typing indicator:', error);
      this.triggerEvent(EVENT_TYPES.ERROR, {
        message: 'Failed to send typing indicator',
        error
      });
      return false;
    }
  }
  
  /**
   * Check if WebSocket is connected
   * @returns {boolean} Connection status
   */
  isConnected() {
    return this.connectionState === CONNECTION_STATES.CONNECTED && 
           this.socket && 
           this.socket.readyState === WebSocket.OPEN;
  }
  
  /**
   * Get current connection state
   * @returns {string} Connection state
   */
  getConnectionState() {
    return this.connectionState;
  }
  
  /**
   * Reset the WebSocket connection state
   * This should be called when the user logs out
   */
  reset() {
    this.disconnect();
    
    // Clear all event handlers
    Object.values(EVENT_TYPES).forEach(eventType => {
      this.eventRegistry[eventType] = [];
    });
    
    this.currentToken = null;
    this.reconnectAttempts = 0;
  }
}

// Create singleton instance
const websocketService = new WebSocketService();

// Export EVENT_TYPES for consumers to use
export { EVENT_TYPES };

export default websocketService; 