import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useChat } from '../../context/ChatContext';
import ConversationsList from './ConversationsList';
import UsersList from './UsersList';
import ChatWindow from './ChatWindow';
import MessageInput from './MessageInput';

const Messaging = () => {
  const { user } = useAuth();
  const { 
    conversations, 
    messages,
    users,
    loading, 
    error,
    isUserTyping,
    isConnected,
    selectedConversation,
    currentMessages,
    handleSelectConversation,
    handleSelectUser,
    formatConversations,
    sendMessage,
    sendTyping
  } = useChat();
  
  const [activeTab, setActiveTab] = useState('conversations'); // 'conversations' or 'users'
  
  // Check if the selected contact is typing
  const isContactTyping = selectedConversation ? 
    isUserTyping(selectedConversation.id) : false;
  
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
  
  // Wrapper for handleSelectUser to also switch to conversations tab
  const handleSelectUserAndSwitchTab = (user) => {
    handleSelectUser(user);
    setActiveTab('conversations');
  };
  
  // Get formatted conversations from context
  const formattedConversations = formatConversations();

  return (
    <div className="messaging-container">
      <div className="sidebar">
        {loading && <div className="loading-indicator">Loading...</div>}
        {error && <div className="error-message">{error}</div>}
        
        {/* WebSocket Connection Status */}
        <div className={`websocket-status ${isConnected ? 'connected' : 'disconnected'}`}>
          {isConnected ? 'Connected' : 'Disconnected'}
        </div>
        
        <div className="tabs">
          <button 
            className={`tab ${activeTab === 'conversations' ? 'active' : ''}`}
            onClick={() => setActiveTab('conversations')}
          >
            Conversations
          </button>
          <button 
            className={`tab ${activeTab === 'users' ? 'active' : ''}`}
            onClick={() => setActiveTab('users')}
          >
            Users
          </button>
        </div>
        
        {activeTab === 'conversations' ? (
          <ConversationsList
            conversations={formattedConversations}
            onSelectConversation={handleSelectConversation}
            selectedId={selectedConversation?.id}
          />
        ) : (
          <UsersList
            users={users}
            onSelectUser={handleSelectUserAndSwitchTab}
          />
        )}
      </div>
      <div className="chat-section">
        {selectedConversation ? (
          <>
            <div className="chat-contact-header">
              <div className="contact-avatar">
                {selectedConversation.contact && (
                  <>
                    <img 
                      src={selectedConversation.contact.avatar} 
                      alt={selectedConversation.contact.name || 'User'} 
                    />
                    <span className="online-indicator"></span>
                  </>
                )}
              </div>
              <div className="contact-info">
                <h3>{selectedConversation.contact?.name || 'Unknown User'}</h3>
                <p className="contact-status">{selectedConversation.contact?.status || 'Status unknown'}</p>
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