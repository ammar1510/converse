import { API_URL, WS_URL } from '../config';
import { getToken } from '../utils/tokenStorage';

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
    this.isConnecting = false;
    this.currentToken = null;
    this.tokenRefreshInProgress = false;
    
    // Hardcode the correct WebSocket URL
    this.wsUrl = 'ws://localhost:8080/api/ws';
    console.log('WebSocket URL base (hardcoded):', this.wsUrl);
  }
  
  /**
   * Connect to the WebSocket server
   * @param {string} token - JWT authentication token
   */
  connect(token) {
    if (!token) {
      console.error('Cannot connect to WebSocket: No token provided');
      return;
    }
    
    if (this.socket || this.isConnecting) {
      console.log('WebSocket already connected or connecting');
      
      // If the token has changed, disconnect and reconnect with the new token
      if (token !== this.currentToken) {
        console.log('Token has changed, reconnecting with new token');
        this.disconnect();
      } else {
        return;
      }
    }
    
    this.currentToken = token;
    this.isConnecting = true;
    console.log('Initializing WebSocket connection');
    
    try {
      // Add token as URL parameter for authentication
      const wsUrlWithToken = `${this.wsUrl}?token=${encodeURIComponent(token)}`;
      console.log('Connecting to WebSocket URL with token parameter:', wsUrlWithToken);
      
      // Create WebSocket connection
      this.socket = new WebSocket(wsUrlWithToken);
      console.log('WebSocket created with URL parameter');
      
      this.socket.onopen = () => {
        console.log('WebSocket connection opened');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        
        // Send authentication message as fallback
        // This is for backward compatibility with servers that expect auth message
        this.socket.send(JSON.stringify({
          type: 'auth',
          token: token
        }));
        console.log('Sent authentication message');
        
        // Dispatch connect event
        this.dispatchEvent('connect', {});
      };
      
      this.socket.onmessage = (event) => {
        console.log('WebSocket message received:', event.data);
        try {
          const data = JSON.parse(event.data);
          console.log('Parsed WebSocket message:', data);
          this.dispatchEvent(data.type || 'message', data);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };
      
      this.socket.onclose = (event) => {
        console.log(`WebSocket disconnected: ${event.code} ${event.reason}`);
        const wasConnected = this.socket !== null;
        this.socket = null;
        this.isConnecting = false;
        
        // Dispatch disconnect event
        this.dispatchEvent('disconnect', { code: event.code, reason: event.reason });
        
        // Don't attempt to reconnect if we're intentionally disconnecting (code 1000)
        // or if we've reached max reconnect attempts
        if (event.code === 1000) {
          console.log('Clean disconnect, not attempting to reconnect');
          return;
        }
        
        // For authentication errors (1006 often indicates auth issues) or other errors,
        // try to refresh the token and reconnect
        console.log(`Abnormal close (${event.code}), attempting to refresh connection`);
        
        // Only try to refresh if we were previously connected
        // This prevents infinite reconnection loops
        if (wasConnected || this.reconnectAttempts < this.maxReconnectAttempts) {
          this.refreshConnection();
        } else {
          console.log('Max reconnect attempts reached, giving up');
        }
      };
      
      this.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.dispatchEvent('error', { error });
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.isConnecting = false;
      this.attemptReconnect(token);
    }
  }
  
  /**
   * Attempt to refresh the connection with a new token
   */
  refreshConnection() {
    if (this.tokenRefreshInProgress) {
      return;
    }
    
    this.tokenRefreshInProgress = true;
    
    // Get the latest token from storage
    const freshToken = getToken();
    
    console.log('Token refresh attempt:', { 
      hasCurrentToken: !!this.currentToken,
      hasFreshToken: !!freshToken,
      tokensMatch: this.currentToken === freshToken,
      tokenLength: freshToken ? freshToken.length : 0
    });
    
    if (freshToken) {
      // Always reconnect with the token from storage, even if it appears to be the same
      // This handles cases where the token might be the same string but invalid on the server
      console.log('Refreshing WebSocket connection with token from storage');
      this.currentToken = freshToken;
      
      // Force disconnect before reconnecting
      this.disconnect();
      
      // Small delay to ensure disconnect completes
      setTimeout(() => {
        this.connect(freshToken);
        this.tokenRefreshInProgress = false;
      }, 500);
    } else {
      console.log('No token available for refresh - user may be logged out');
      this.tokenRefreshInProgress = false;
    }
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
    
    this.isConnecting = false;
    this.currentToken = null;
    this.reconnectAttempts = 0;
  }
  
  /**
   * Attempt to reconnect to the WebSocket server
   * @param {string} token - JWT authentication token
   */
  attemptReconnect(token) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached');
      this.dispatchEvent('reconnect_failed', {});
      return;
    }
    
    // Exponential backoff with jitter
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    const jitter = Math.random() * 1000;
    this.reconnectAttempts++;
    
    console.log(`Attempting to reconnect in ${Math.floor((delay + jitter) / 1000)}s...`);
    
    this.reconnectTimeout = setTimeout(() => {
      this.connect(token);
    }, delay + jitter);
    
    this.dispatchEvent('reconnecting', { attempt: this.reconnectAttempts });
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
    console.log(`Dispatching ${event} event:`, data);
    if (this.listeners[event]) {
      console.log(`Found ${this.listeners[event].length} listeners for ${event} event`);
      this.listeners[event].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in ${event} event handler:`, error);
        }
      });
    } else {
      console.log(`No listeners found for ${event} event`);
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
      console.error('WebSocket not connected');
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
    return this.socket && this.socket.readyState === WebSocket.OPEN;
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
    this.tokenRefreshInProgress = false;
  }
}

// Create singleton instance
const websocketService = new WebSocketService();

export default websocketService; 