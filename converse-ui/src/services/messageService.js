import api from './api';

/**
 * Simple cache implementation for API responses
 */
class ApiCache {
  constructor(ttl = 60000) { // Default TTL: 1 minute
    this.cache = new Map();
    this.ttl = ttl;
  }

  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;
    
    // Check if the cache item has expired
    if (Date.now() > item.expiry) {
      this.cache.delete(key);
      return null;
    }
    
    return item.data;
  }

  set(key, data) {
    const expiry = Date.now() + this.ttl;
    this.cache.set(key, { data, expiry });
  }

  invalidate(key) {
    if (key) {
      this.cache.delete(key);
    } else {
      this.cache.clear();
    }
  }
}

// Create cache instances with different TTLs for different types of data
const usersCache = new ApiCache(300000); // 5 minutes for users
const messagesCache = new ApiCache(15000); // 15 seconds for messages

/**
 * Service for handling message-related API calls
 */
export const messageService = {
  /**
   * Get all users except the current user
   * @returns {Promise} Promise with users data
   */
  getAllUsers: async () => {
    // Check cache first
    const cachedUsers = usersCache.get('users');
    if (cachedUsers) {
      console.log('Using cached users data');
      return cachedUsers;
    }
    
    const response = await api.get('/users');
    
    // Cache the response
    usersCache.set('users', response.data);
    
    return response.data;
  },

  /**
   * Get all messages for the current user
   * @returns {Promise} Promise with messages data
   */
  getMessages: async () => {
    // Check cache first
    const cachedMessages = messagesCache.get('all_messages');
    if (cachedMessages) {
      console.log('Using cached messages data');
      return cachedMessages;
    }
    
    const response = await api.get('/messages');
    
    // Cache the response
    messagesCache.set('all_messages', response.data);
    
    return response.data;
  },
  
  /**
   * Get conversation between current user and another user
   * @param {string} otherUserId - UUID of the other user
   * @returns {Promise} Promise with conversation messages
   */
  getConversation: async (otherUserId) => {
    if (!otherUserId) {
      throw new Error('otherUserId is required');
    }
    
    // Create a unique cache key for this conversation
    const cacheKey = `conversation_${otherUserId}`;
    
    // Check cache first
    const cachedConversation = messagesCache.get(cacheKey);
    if (cachedConversation) {
      console.log(`Using cached conversation data with user ${otherUserId}`);
      return cachedConversation;
    }
    
    const response = await api.get(`/messages/conversation/${otherUserId}`);
    
    // Cache the response
    messagesCache.set(cacheKey, response.data);
    
    return response.data;
  },
  
  /**
   * Send a new message
   * @param {string} receiverId - UUID of the message recipient
   * @param {string} content - Message content
   * @returns {Promise} Promise with the created message
   */
  sendMessage: async (receiverId, content) => {
    const response = await api.post('/messages', {
      receiver_id: receiverId,
      content: content
    });
    
    // Invalidate related caches when sending a message
    messagesCache.invalidate('all_messages');
    messagesCache.invalidate(`conversation_${receiverId}`);
    
    return response.data;
  },
  
  /**
   * Mark a message as read
   * @param {string} messageId - UUID of the message
   * @returns {Promise} Promise with operation result
   */
  markAsRead: async (messageId) => {
    const response = await api.put(`/messages/${messageId}/read`);
    
    // Since we don't know which conversation this belongs to,
    // we'll invalidate all message caches
    messagesCache.invalidate();
    
    return response.data;
  },
  
  /**
   * Invalidate all cached data
   * Used when user logs out or when data should be refreshed
   */
  invalidateCache: () => {
    usersCache.invalidate();
    messagesCache.invalidate();
  },
  
  /**
   * Process raw messages into conversations
   * @param {Array} messages - Raw messages from API
   * @param {string} currentUserId - UUID of current user
   * @returns {Array} Processed conversations
   */
  processConversations: (messages, currentUserId) => {
    // Group messages by the other user (sender or receiver)
    const conversationsMap = {};
    
    // Add null check to prevent "Cannot read properties of null" error
    if (!messages || !Array.isArray(messages)) {
      console.warn('No messages to process or messages is not an array');
      return [];
    }
    
    messages.forEach(message => {
      // Determine the other user in the conversation
      const otherUserId = message.sender_id === currentUserId 
        ? message.receiver_id 
        : message.sender_id;
      
      if (!conversationsMap[otherUserId]) {
        conversationsMap[otherUserId] = {
          id: otherUserId,
          messages: [],
          lastMessage: null,
          lastMessageTime: null
        };
      }
      
      // Add message to conversation
      conversationsMap[otherUserId].messages.push(message);
      
      // Update last message if this is newer
      if (!conversationsMap[otherUserId].lastMessageTime || 
          new Date(message.created_at) > new Date(conversationsMap[otherUserId].lastMessageTime)) {
        conversationsMap[otherUserId].lastMessage = message;
        conversationsMap[otherUserId].lastMessageTime = message.created_at;
      }
    });
    
    // Convert map to array and sort by last message time
    return Object.values(conversationsMap)
      .sort((a, b) => new Date(b.lastMessageTime) - new Date(a.lastMessageTime));
  }
};

export default messageService; 