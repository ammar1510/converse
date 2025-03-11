import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useChat } from '../../context/ChatContext';
import ConversationsList from './ConversationsList';
import ChatWindow from './ChatWindow';
import MessageInput from './MessageInput';

const Messaging = () => {
  const { user } = useAuth();
  const { 
    conversations, 
    messages, 
    loading, 
    error,
    isUserTyping,
    fetchConversations, 
    fetchConversation,
    sendMessage,
    sendTyping
  } = useChat();
  
  const [selectedConversation, setSelectedConversation] = useState(null);
  const [currentMessages, setCurrentMessages] = useState([]);
  
  // Fetch conversations on component mount
  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);
  
  // Update selected conversation when conversations change
  useEffect(() => {
    if (!selectedConversation && conversations.length > 0) {
      handleSelectConversation(conversations[0]);
    }
  }, [conversations, selectedConversation]);
  
  // Update current messages when selected conversation changes
  useEffect(() => {
    if (selectedConversation) {
      const conversationMessages = messages[selectedConversation.id] || [];
      setCurrentMessages(conversationMessages);
      
      // Fetch conversation messages if not already loaded
      if (conversationMessages.length === 0) {
        fetchConversation(selectedConversation.id);
      }
    }
  }, [selectedConversation, messages, fetchConversation]);

  const handleSelectConversation = (conversation) => {
    setSelectedConversation(conversation);
  };

  const handleSendMessage = async (text) => {
    if (!selectedConversation) return;
    
    try {
      await sendMessage(selectedConversation.id, text);
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  };
  
  const handleTyping = (isTyping) => {
    if (!selectedConversation) return;
    sendTyping(selectedConversation.id, isTyping);
  };
  
  // Check if the selected contact is typing
  const isContactTyping = selectedConversation ? 
    isUserTyping(selectedConversation.id) : false;
  
  // Format conversations for display
  const formattedConversations = conversations.map(convo => {
    // Get the last message
    const lastMsg = convo.lastMessage || {};
    
    // Format timestamp
    const timestamp = lastMsg.created_at 
      ? new Date(lastMsg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : '';
    
    // Get contact info from sender or receiver
    const contact = {
      id: convo.id,
      name: 'Unknown User', // This should be replaced with actual user data
      status: 'Online', // This should be replaced with actual status
      avatar: `https://ui-avatars.com/api/?name=${convo.id}&background=4ead7c&color=fff&rounded=true&size=128`
    };
    
    return {
      id: convo.id,
      title: `Conversation with ${contact.name}`,
      lastMessage: lastMsg.content || 'No messages yet',
      timestamp,
      contact
    };
  });

  return (
    <div className="messaging-container">
      <div className="sidebar">
        {loading && <div className="loading-indicator">Loading conversations...</div>}
        {error && <div className="error-message">{error}</div>}
        
        <ConversationsList
          conversations={formattedConversations}
          onSelectConversation={handleSelectConversation}
          selectedId={selectedConversation?.id}
        />
      </div>
      <div className="chat-section">
        {selectedConversation ? (
          <>
            <div className="chat-contact-header">
              <div className="contact-avatar">
                <img src={selectedConversation.contact.avatar} alt={selectedConversation.contact.name} />
                <span className="online-indicator"></span>
              </div>
              <div className="contact-info">
                <h3>{selectedConversation.contact.name}</h3>
                <p className="contact-status">{selectedConversation.contact.status}</p>
              </div>
            </div>
            <div className="chat-messages-container">
              {loading && <div className="loading-indicator">Loading messages...</div>}
              {error && <div className="error-message">{error}</div>}
              
              <ChatWindow 
                messages={currentMessages} 
                currentUser={user} 
              />
              
              {/* Typing indicator - shown when contact is typing */}
              {isContactTyping && (
                <div className="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              )}
            </div>
            <MessageInput 
              onSendMessage={handleSendMessage} 
              onTyping={handleTyping}
            />
          </>
        ) : (
          <p>Please select a conversation</p>
        )}
      </div>
    </div>
  );
};

export default Messaging; 