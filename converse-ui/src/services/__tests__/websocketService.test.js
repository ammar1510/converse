import websocketService from '../websocketService';

// Mock WebSocket
class MockWebSocket {
  constructor(url, protocols) {
    this.url = url;
    this.protocols = protocols;
    this.readyState = 0; // CONNECTING
    this.OPEN = 1;
    
    // Store the event handlers
    this.onopen = null;
    this.onmessage = null;
    this.onclose = null;
    this.onerror = null;
    
    // Mock methods
    this.send = jest.fn();
    this.close = jest.fn();
    this.setRequestHeader = jest.fn();
    
    // Simulate connection
    setTimeout(() => {
      this.readyState = 1; // OPEN
      if (this.onopen) this.onopen();
    }, 0);
  }
}

// Mock global WebSocket
global.WebSocket = MockWebSocket;

describe('WebSocketService', () => {
  beforeEach(() => {
    // Reset the service before each test
    websocketService.disconnect();
    jest.clearAllMocks();
  });
  
  test('should connect with JWT token in protocol', () => {
    const token = 'test-jwt-token';
    websocketService.connect(token);
    
    // Wait for the connection to be established
    return new Promise(resolve => {
      setTimeout(() => {
        // Check if WebSocket was created with the correct protocol
        expect(websocketService.socket.protocols).toEqual(['jwt', token]);
        
        // Check if auth message was sent
        expect(websocketService.socket.send).toHaveBeenCalledWith(
          expect.stringContaining('"type":"auth"')
        );
        expect(websocketService.socket.send).toHaveBeenCalledWith(
          expect.stringContaining(`"token":"${token}"`)
        );
        
        resolve();
      }, 10);
    });
  });
  
  test('should attempt to set Authorization header if supported', () => {
    const token = 'test-jwt-token';
    websocketService.connect(token);
    
    // Wait for the connection to be established
    return new Promise(resolve => {
      setTimeout(() => {
        // Check if setRequestHeader was called with the correct Authorization header
        expect(websocketService.socket.setRequestHeader).toHaveBeenCalledWith(
          'Authorization', 
          `Bearer ${token}`
        );
        
        resolve();
      }, 10);
    });
  });
  
  test('should send messages with correct format', () => {
    const token = 'test-jwt-token';
    const receiverId = '123e4567-e89b-12d3-a456-426614174000';
    const content = 'Hello, world!';
    
    websocketService.connect(token);
    
    // Wait for the connection to be established
    return new Promise(resolve => {
      setTimeout(() => {
        // Reset the mock to clear the auth message
        websocketService.socket.send.mockClear();
        
        // Send a message
        websocketService.sendMessage(receiverId, content);
        
        // Check if the message was sent with the correct format
        expect(websocketService.socket.send).toHaveBeenCalledWith(
          expect.stringContaining('"type":"message"')
        );
        expect(websocketService.socket.send).toHaveBeenCalledWith(
          expect.stringContaining(`"receiver_id":"${receiverId}"`)
        );
        expect(websocketService.socket.send).toHaveBeenCalledWith(
          expect.stringContaining(`"content":"${content}"`)
        );
        
        resolve();
      }, 10);
    });
  });
  
  test('should handle reconnection with token', () => {
    const token = 'test-jwt-token';
    websocketService.connect(token);
    
    // Wait for the connection to be established
    return new Promise(resolve => {
      setTimeout(() => {
        // Simulate connection close
        websocketService.socket.onclose({ code: 1006, reason: 'Connection lost' });
        
        // Check if reconnection was attempted
        expect(websocketService.reconnectAttempts).toBe(1);
        
        // Fast-forward timers to trigger reconnect
        jest.advanceTimersByTime(2000);
        
        // Check if a new connection was established with the token
        expect(websocketService.socket.protocols).toEqual(['jwt', token]);
        
        resolve();
      }, 10);
    });
  });
}); 