import { API_URL } from '../config';

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
    
    // Convert HTTP/HTTPS to WS/WSS
    const wsProtocol = API_URL.startsWith('https') ? 'wss' : 'ws';
    this.wsUrl = API_URL.replace(/^https?:\/\//, `${wsProtocol}://`) + '/ws';
  }
  
  /**
   * Connect to the WebSocket server
   * @param {string} token - JWT authentication token
   */
  connect(token) {
    if (this.socket || this.isConnecting) {
      return;
    }
    
    this.isConnecting = true;
    
    try {
      // Add token to WebSocket URL as a query parameter for servers that don't support custom headers
      // AND include it in the protocol for browsers that support it
      const wsUrlWithToken = `${this.wsUrl}`;
      
      // Create WebSocket with token in protocol (supported by most browsers)
      this.socket = new WebSocket(wsUrlWithToken, ['jwt', token]);
      
      // Set up custom headers for the WebSocket connection
      // Note: This is only supported in some environments and may not work in all browsers
      if (this.socket.setRequestHeader) {
        this.socket.setRequestHeader('Authorization', `Bearer ${token}`);
      }
      
      this.socket.onopen = () => {
        console.log('WebSocket connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        
        // Send authentication message as fallback
        // This is for backward compatibility with servers that expect auth message
        this.socket.send(JSON.stringify({
          type: 'auth',
          token: token
        }));
        
        // Dispatch connect event
        this.dispatchEvent('connect', {});
      };
      
      this.socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.dispatchEvent(data.type || 'message', data);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };
      
      this.socket.onclose = (event) => {
        console.log(`WebSocket disconnected: ${event.code} ${event.reason}`);
        this.socket = null;
        this.isConnecting = false;
        this.dispatchEvent('disconnect', { code: event.code, reason: event.reason });
        
        // Attempt to reconnect if not closed cleanly
        if (event.code !== 1000) {
          this.attemptReconnect(token);
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
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in ${event} event handler:`, error);
        }
      });
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
}

// Create singleton instance
const websocketService = new WebSocketService();

export default websocketService; 