import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react';
import { useAuth } from './AuthContext';
import websocketService, { EVENT_TYPES } from '../services/websocketService';
import messageService from '../services/messageService';
import { getToken } from '../utils/tokenStorage';
import { formatTimestamp, generateAvatarUrl, registerEventHandlers } from '../utils/utils';

// Create context
const ChatContext = createContext();

/**
 * ChatProvider component to manage chat state and WebSocket events
 */
export const ChatProvider = ({ children }) => {
  const { isAuthenticated, user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState({});
  const [conversations, setConversations] = useState([]);
  const [users, setUsers] = useState([]);
  const [typingUsers, setTypingUsers] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedConversation, setSelectedConversation] = useState(null);
  const [currentMessages, setCurrentMessages] = useState([]);
  
  // Define a typingTimeouts ref at the component top level
  const typingTimeoutsRef = useRef({});
  
  // Get token directly from storage to ensure it's always current
  const token = getToken();

  /**
   * Format conversations for display
   * @returns {array} Formatted conversations
   */
  const formatConversations = useCallback(() => {
    if (!user?.id) return [];
    
    return conversations.map(convo => {
      // Get the last message
      const lastMsg = convo.lastMessage || {};
      
      // Format timestamp
      const timestamp = lastMsg.created_at 
        ? formatTimestamp(lastMsg.created_at)
        : '';
      
      // Find the user in the users array
      const contactUser = users.find(u => u.id === convo.id);
      
      // Get contact info from sender or receiver
      const contact = {
        id: convo.id,
        name: contactUser ? (contactUser.display_name || contactUser.username) : 'Unknown User',
        status: 'Online', // This should be replaced with actual status
        avatar: contactUser?.avatar_url || generateAvatarUrl(contactUser?.username || convo.id)
      };
      
      return {
        id: convo.id,
        title: `Conversation with ${contact.name}`,
        lastMessage: lastMsg.content || 'No messages yet',
        timestamp,
        contact
      };
    });
  }, [conversations, users, user?.id]);

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
        // Sort messages by timestamp (oldest first)
        const sortedMessages = [...convo.messages].sort(
          (a, b) => new Date(a.created_at) - new Date(b.created_at)
        );
        messagesByConversation[convo.id] = sortedMessages;
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
      
      // Sort messages by timestamp (oldest first)
      const sortedConversationData = [...conversationData].sort(
        (a, b) => new Date(a.created_at) - new Date(b.created_at)
      );
      
      setMessages(prev => ({
        ...prev,
        [userId]: sortedConversationData
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
   * Handle selecting a conversation
   * @param {object} conversation - The conversation to select
   */
  const handleSelectConversation = useCallback((conversation) => {
    setSelectedConversation(conversation);
    
    if (conversation) {
      const conversationMessages = messages[conversation.id] || [];
      setCurrentMessages(conversationMessages);
      
      // Only fetch if no messages loaded
      if (conversationMessages.length === 0) {
        fetchConversation(conversation.id);
      }
    }
  }, [messages, fetchConversation]);

  /**
   * Handle selecting a user to start or continue a conversation
   * @param {object} selectedUser - The user to start/continue conversation with
   */
  const handleSelectUser = useCallback((selectedUser) => {
    // Create a conversation object from the selected user
    const conversation = {
      id: selectedUser.id,
      title: `Conversation with ${selectedUser.display_name || selectedUser.username}`,
      lastMessage: 'Start a new conversation',
      timestamp: '',
      contact: {
        id: selectedUser.id,
        name: selectedUser.display_name || selectedUser.username,
        status: 'Online', // Default status
        avatar: selectedUser.avatar_url || generateAvatarUrl(selectedUser.username)
      }
    };
    
    handleSelectConversation(conversation);
  }, [handleSelectConversation]);
  
  // Update selected conversation when conversations change
  useEffect(() => {
    if (!selectedConversation && conversations.length > 0) {
      // Format the first conversation and select it if none is selected
      const formattedConversations = formatConversations();
      if (formattedConversations.length > 0) {
        setSelectedConversation(formattedConversations[0]);
      }
    }
  }, [conversations, selectedConversation, formatConversations]);
  
  // Update current messages when selected conversation changes
  useEffect(() => {
    if (selectedConversation) {
      const conversationMessages = messages[selectedConversation.id] || [];
      setCurrentMessages(conversationMessages);
    }
  }, [selectedConversation, messages]);

  // Periodic refresh of current conversation
  useEffect(() => {
    if (!selectedConversation) return;
    
    const refreshInterval = setInterval(() => {
      fetchConversation(selectedConversation.id).catch(() => {
        // Silently handle errors during background refresh
      });
    }, 15000); // Reduced frequency
    
    return () => clearInterval(refreshInterval);
  }, [selectedConversation, fetchConversation]);
  
  // Setup WebSocket event listeners
  useEffect(() => {
    // Always get the latest token from storage
    const currentToken = getToken();
    
    // Define event handlers
    const handleConnect = () => {
      setIsConnected(true);
      setError(null);
      console.log('WebSocket connected successfully, will now fetch conversations');
      // Fetch conversations whenever we connect to ensure latest data
      fetchConversations();
    };
    
    const handleDisconnect = () => {
      setIsConnected(false);
      console.log('WebSocket disconnected');
    };
    
    const handleMessage = (data) => {
      console.log('Received message in chat context:', data);
      // Add new message to state
      setMessages(prev => {
        const conversationId = data.sender_id === user?.id ? data.receiver_id : data.sender_id;
        const conversationMessages = [...(prev[conversationId] || [])];
        
        // Check if message already exists to avoid duplicates
        if (!conversationMessages.some(msg => msg.id === data.id)) {
          conversationMessages.push(data);
          // Sort messages by timestamp (oldest first)
          conversationMessages.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
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
      console.error('WebSocket error in chat context:', data);
      setError(`WebSocket error: ${data.message || 'Unknown error'}`);
    };
    
    if (isAuthenticated && currentToken && user?.id) {
      console.log('User is authenticated. Setting up WebSocket connection...');
      
      // Disconnect any existing connection first
      websocketService.disconnect();
      
      // Register event handlers using the utility function
      const cleanup = registerEventHandlers(websocketService, {
        [EVENT_TYPES.CONNECT]: handleConnect,
        [EVENT_TYPES.DISCONNECT]: handleDisconnect,
        [EVENT_TYPES.MESSAGE]: handleMessage,
        [EVENT_TYPES.TYPING]: handleTyping,
        [EVENT_TYPES.ERROR]: handleError
      });
      
      // Connect to WebSocket with current token
      console.log('Connecting to WebSocket with token');
      websocketService.connect(currentToken);
      
      // Clean up on unmount or when dependencies change
      return cleanup;
    } else if (!isAuthenticated) {
      console.log('User is not authenticated, disconnecting WebSocket');
      // Ensure WebSocket is disconnected when not authenticated
      websocketService.disconnect();
      setIsConnected(false);
    }
  }, [isAuthenticated, user?.id, fetchConversations, updateConversationWithMessage]);
  
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
        // Sort messages by timestamp (oldest first)
        conversationMessages.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
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
  
  /**
   * Fetch all users from the API
   */
  const fetchUsers = useCallback(async () => {
    if (!isAuthenticated) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const usersData = await messageService.getAllUsers();
      setUsers(usersData);
    } catch (err) {
      console.error('Error fetching users:', err);
      setError('Failed to load users. Please try again.');
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  // Fetch users when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      fetchUsers();
    }
  }, [isAuthenticated, fetchUsers]);
  
  // Add new utility function for managing typing indicators with timeouts
  const handleTypingIndicator = useCallback((conversationId, isTyping) => {
    if (!conversationId) return;
    
    if (isTyping) {
      // Send typing indicator
      sendTyping(conversationId, true);
      
      // Clear any existing timeout for this user
      if (typingTimeoutsRef.current[conversationId]) {
        clearTimeout(typingTimeoutsRef.current[conversationId]);
      }
      
      // Set timeout to automatically clear typing indicator after inactivity
      typingTimeoutsRef.current[conversationId] = setTimeout(() => {
        sendTyping(conversationId, false);
        delete typingTimeoutsRef.current[conversationId];
      }, 3000); // 3 seconds of inactivity
    } else {
      // Turn off typing indicator
      sendTyping(conversationId, false);
      
      // Clear any existing timeout
      if (typingTimeoutsRef.current[conversationId]) {
        clearTimeout(typingTimeoutsRef.current[conversationId]);
        delete typingTimeoutsRef.current[conversationId];
      }
    }
  }, [sendTyping]);
  
  // Context value
  const value = {
    isConnected,
    loading,
    error,
    conversations,
    messages,
    users,
    typingUsers,
    fetchConversations,
    fetchConversation,
    fetchUsers,
    sendMessage,
    sendTyping,
    isUserTyping,
    selectedConversation,
    currentMessages,
    handleSelectConversation,
    handleSelectUser,
    formatConversations,
    handleTypingIndicator
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