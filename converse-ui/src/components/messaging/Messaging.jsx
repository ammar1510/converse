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
    fetchConversations, 
    fetchConversation,
    fetchUsers,
    sendMessage,
    sendTyping
  } = useChat();
  
  const [selectedConversation, setSelectedConversation] = useState(null);
  const [currentMessages, setCurrentMessages] = useState([]);
  const [activeTab, setActiveTab] = useState('conversations'); // 'conversations' or 'users'
  
  // Fetch conversations and users on component mount
  useEffect(() => {
    fetchConversations();
    fetchUsers();
  }, [fetchConversations, fetchUsers]);
  
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

  const handleSelectUser = (selectedUser) => {
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
        avatar: selectedUser.avatar_url || `https://ui-avatars.com/api/?name=${selectedUser.username}&background=4ead7c&color=fff&rounded=true&size=128`
      }
    };
    
    handleSelectConversation(conversation);
    setActiveTab('conversations');
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
    
    // Find the user in the users array
    const contactUser = users.find(u => u.id === convo.id);
    
    // Get contact info from sender or receiver
    const contact = {
      id: convo.id,
      name: contactUser ? (contactUser.display_name || contactUser.username) : 'Unknown User',
      status: 'Online', // This should be replaced with actual status
      avatar: contactUser?.avatar_url || `https://ui-avatars.com/api/?name=${convo.id}&background=4ead7c&color=fff&rounded=true&size=128`
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
        {loading && <div className="loading-indicator">Loading...</div>}
        {error && <div className="error-message">{error}</div>}
        
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
            onSelectUser={handleSelectUser}
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
                      src={selectedConversation.contact.avatar || `https://ui-avatars.com/api/?name=User&background=4ead7c&color=fff&rounded=true&size=128`} 
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