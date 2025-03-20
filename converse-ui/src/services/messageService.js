import api from './api';

/**
 * Service for handling message-related API calls
 */
export const messageService = {
  /**
   * Get all users except the current user
   * @returns {Promise} Promise with users data
   */
  getAllUsers: async () => {
    const response = await api.get('/users');
    return response.data;
  },

  /**
   * Get all messages for the current user
   * @returns {Promise} Promise with messages data
   */
  getMessages: async () => {
    const response = await api.get('/messages');
    return response.data;
  },
  
  /**
   * Get conversation between current user and another user
   * @param {string} otherUserId - UUID of the other user
   * @returns {Promise} Promise with conversation messages
   */
  getConversation: async (otherUserId) => {
    const response = await api.get(`/messages/conversation/${otherUserId}`);
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
    return response.data;
  },
  
  /**
   * Mark a message as read
   * @param {string} messageId - UUID of the message
   * @returns {Promise} Promise with operation result
   */
  markAsRead: async (messageId) => {
    const response = await api.put(`/messages/${messageId}/read`);
    return response.data;
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