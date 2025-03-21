import { WS_URL } from '../config';
import { getToken } from '../utils/tokenStorage';

// Connection states for state machine
const CONNECTION_STATES = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting'
};

/**
 * Service for managing WebSocket connections and real-time messaging
 */
class WebSocketService {
  constructor() {
    this.socket = null;
    this.listeners = {};
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectTimeout = null;
    this.currentToken = null;
    
    // State machine for connection status
    this.connectionState = CONNECTION_STATES.DISCONNECTED;
    
    // Use the URL from config
    this.wsUrl = WS_URL;
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
      this.dispatchEvent('error', { message: 'No token provided' });
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
      
      // Dispatch connect event
      this.dispatchEvent('connect', {});
    };
    
    // Message received
    this.socket.onmessage = (event) => {
      console.log('WebSocket message received');
      try {
        const data = JSON.parse(event.data);
        this.dispatchEvent(data.type || 'message', data);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
        this.dispatchEvent('error', { message: 'Failed to parse message', error });
      }
    };
    
    // Connection closed
    this.socket.onclose = (event) => {
      console.log(`WebSocket disconnected: ${event.code} ${event.reason}`);
      const wasConnected = this.connectionState === CONNECTION_STATES.CONNECTED;
      this.socket = null;
      
      // Update state
      this.connectionState = CONNECTION_STATES.DISCONNECTED;
      
      // Dispatch disconnect event
      this.dispatchEvent('disconnect', { code: event.code, reason: event.reason });
      
      // Handle reconnection based on close code
      this.handleCloseEvent(event, wasConnected);
    };
    
    // Error occurred
    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.dispatchEvent('error', { error });
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
      this.dispatchEvent('reconnect_failed', { attempts: this.reconnectAttempts });
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
      this.dispatchEvent('reconnect_failed', { attempts: this.reconnectAttempts });
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
    
    this.dispatchEvent('reconnecting', { 
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
   * Add event listener
   * @param {string} event - Event name
   * @param {function} callback - Event callback
   */
  addEventListener(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }
  
  /**
   * Remove event listener
   * @param {string} event - Event name
   * @param {function} callback - Event callback to remove
   */
  removeEventListener(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback);
    }
  }
  
  /**
   * Dispatch event to all listeners
   * @param {string} event - Event name
   * @param {object} data - Event data
   */
  dispatchEvent(event, data) {
    if (this.listeners[event] && this.listeners[event].length > 0) {
      console.log(`Dispatching ${event} event to ${this.listeners[event].length} listeners`);
      this.listeners[event].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in ${event} event handler:`, error);
        }
      });
    } else {
      console.log(`No listeners for ${event} event`);
    }
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
    this.listeners = {};
    this.currentToken = null;
    this.reconnectAttempts = 0;
  }
}

// Create singleton instance
const websocketService = new WebSocketService();

export default websocketService; 