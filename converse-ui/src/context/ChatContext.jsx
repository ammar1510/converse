import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useAuth } from './AuthContext';
import websocketService from '../services/websocketService';
import messageService from '../services/messageService';

// Create context
const ChatContext = createContext();

/**
 * ChatProvider component to manage chat state and WebSocket events
 */
export const ChatProvider = ({ children }) => {
  const { isAuthenticated, token, user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState({});
  const [conversations, setConversations] = useState([]);
  const [typingUsers, setTypingUsers] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Connect to WebSocket when authenticated
  useEffect(() => {
    if (isAuthenticated && token) {
      // Connect to WebSocket
      websocketService.connect(token);
      
      // Set up event listeners
      const handleConnect = () => {
        setIsConnected(true);
        setError(null);
      };
      
      const handleDisconnect = () => {
        setIsConnected(false);
      };
      
      const handleMessage = (data) => {
        // Add new message to state
        setMessages(prev => {
          const conversationId = data.sender_id === user?.id ? data.receiver_id : data.sender_id;
          const conversationMessages = [...(prev[conversationId] || [])];
          
          // Check if message already exists to avoid duplicates
          if (!conversationMessages.some(msg => msg.id === data.id)) {
            conversationMessages.push(data);
          }
          
          return {
            ...prev,
            [conversationId]: conversationMessages
          };
        });
        
        // Update conversations list
        updateConversationWithMessage(data);
      };
      
      const handleTyping = (data) => {
        if (data.is_typing) {
          setTypingUsers(prev => ({
            ...prev,
            [data.sender_id]: Date.now()
          }));
        } else {
          setTypingUsers(prev => {
            const newState = { ...prev };
            delete newState[data.sender_id];
            return newState;
          });
        }
      };
      
      const handleError = (data) => {
        setError(`WebSocket error: ${data.message || 'Unknown error'}`);
      };
      
      // Add event listeners
      websocketService.addEventListener('connect', handleConnect);
      websocketService.addEventListener('disconnect', handleDisconnect);
      websocketService.addEventListener('message', handleMessage);
      websocketService.addEventListener('typing', handleTyping);
      websocketService.addEventListener('error', handleError);
      
      // Fetch initial conversations
      fetchConversations();
      
      // Clean up on unmount
      return () => {
        websocketService.removeEventListener('connect', handleConnect);
        websocketService.removeEventListener('disconnect', handleDisconnect);
        websocketService.removeEventListener('message', handleMessage);
        websocketService.removeEventListener('typing', handleTyping);
        websocketService.removeEventListener('error', handleError);
        
        websocketService.disconnect();
      };
    }
  }, [isAuthenticated, token, user?.id]);
  
  /**
   * Fetch all conversations for the current user
   */
  const fetchConversations = useCallback(async () => {
    if (!user?.id) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const messagesData = await messageService.getMessages();
      const processedConversations = messageService.processConversations(messagesData, user.id);
      
      setConversations(processedConversations);
      
      // Organize messages by conversation
      const messagesByConversation = {};
      processedConversations.forEach(convo => {
        messagesByConversation[convo.id] = convo.messages;
      });
      
      setMessages(messagesByConversation);
    } catch (err) {
      console.error('Error fetching conversations:', err);
      setError('Failed to load conversations');
    } finally {
      setLoading(false);
    }
  }, [user?.id]);
  
  /**
   * Fetch messages for a specific conversation
   * @param {string} userId - UUID of the other user in the conversation
   */
  const fetchConversation = useCallback(async (userId) => {
    if (!user?.id) return;
    
    try {
      setLoading(true);
      setError(null);
      
      const conversationData = await messageService.getConversation(userId);
      
      setMessages(prev => ({
        ...prev,
        [userId]: conversationData
      }));
      
      // Mark unread messages as read
      const unreadMessages = conversationData.filter(
        msg => !msg.is_read && msg.sender_id === userId
      );
      
      for (const msg of unreadMessages) {
        await messageService.markAsRead(msg.id);
      }
      
      return conversationData;
    } catch (err) {
      console.error(`Error fetching conversation with ${userId}:`, err);
      setError('Failed to load conversation');
      return [];
    } finally {
      setLoading(false);
    }
  }, [user?.id]);
  
  /**
   * Send a message to another user
   * @param {string} receiverId - UUID of the message recipient
   * @param {string} content - Message content
   */
  const sendMessage = useCallback(async (receiverId, content) => {
    if (!user?.id) return null;
    
    try {
      setError(null);
      
      // Send via REST API for persistence
      const newMessage = await messageService.sendMessage(receiverId, content);
      
      // Also try to send via WebSocket for real-time
      websocketService.sendMessage(receiverId, content);
      
      // Update local state
      setMessages(prev => {
        const conversationMessages = [...(prev[receiverId] || []), newMessage];
        return {
          ...prev,
          [receiverId]: conversationMessages
        };
      });
      
      // Update conversations list
      updateConversationWithMessage(newMessage);
      
      return newMessage;
    } catch (err) {
      console.error('Error sending message:', err);
      setError('Failed to send message');
      return null;
    }
  }, [user?.id]);
  
  /**
   * Update conversations list with a new message
   * @param {object} message - The new message
   */
  const updateConversationWithMessage = useCallback((message) => {
    if (!user?.id) return;
    
    setConversations(prev => {
      // Determine the conversation ID (the other user's ID)
      const conversationId = message.sender_id === user.id ? message.receiver_id : message.sender_id;
      
      // Find existing conversation
      const existingIndex = prev.findIndex(c => c.id === conversationId);
      
      // Create a copy of the conversations array
      const updatedConversations = [...prev];
      
      if (existingIndex >= 0) {
        // Update existing conversation
        const conversation = { ...updatedConversations[existingIndex] };
        conversation.lastMessage = message;
        conversation.lastMessageTime = message.created_at;
        conversation.messages = [...(conversation.messages || []), message];
        
        // Remove from current position
        updatedConversations.splice(existingIndex, 1);
        // Add to the beginning (most recent)
        updatedConversations.unshift(conversation);
      } else {
        // Create new conversation
        updatedConversations.unshift({
          id: conversationId,
          messages: [message],
          lastMessage: message,
          lastMessageTime: message.created_at
        });
      }
      
      return updatedConversations;
    });
  }, [user?.id]);
  
  /**
   * Send typing indicator
   * @param {string} receiverId - UUID of the message recipient
   * @param {boolean} isTyping - Whether the user is typing
   */
  const sendTyping = useCallback((receiverId, isTyping) => {
    return websocketService.sendTyping(receiverId, isTyping);
  }, []);
  
  /**
   * Check if a user is currently typing
   * @param {string} userId - UUID of the user to check
   * @returns {boolean} Whether the user is typing
   */
  const isUserTyping = useCallback((userId) => {
    const timestamp = typingUsers[userId];
    if (!timestamp) return false;
    
    // Consider typing indicator valid for 3 seconds
    return Date.now() - timestamp < 3000;
  }, [typingUsers]);
  
  // Context value
  const value = {
    isConnected,
    loading,
    error,
    conversations,
    messages,
    fetchConversations,
    fetchConversation,
    sendMessage,
    sendTyping,
    isUserTyping
  };
  
  return (
    <ChatContext.Provider value={value}>
      {children}
    </ChatContext.Provider>
  );
};

/**
 * Hook to use the chat context
 * @returns {object} Chat context
 */
export const useChat = () => {
  const context = useContext(ChatContext);
  if (context === undefined) {
    throw new Error('useChat must be used within a ChatProvider');
  }
  return context;
};

export default ChatContext; 