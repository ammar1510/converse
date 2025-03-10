import React, { useState, useEffect } from 'react';
import ConversationsList from './ConversationsList';
import ChatWindow from './ChatWindow';
import MessageInput from './MessageInput';

const Messaging = () => {
  const defaultConversations = [
    { 
      id: 1, 
      title: 'Conversation 1', 
      lastMessage: 'Hey, how are you?', 
      timestamp: '10:00 AM',
      contact: {
        name: 'Alice',
        status: 'Online • Last seen just now',
        avatar: 'https://ui-avatars.com/api/?name=Alice&background=4ead7c&color=fff&rounded=true&size=128'
      }
    },
    { 
      id: 2, 
      title: 'Conversation 2', 
      lastMessage: "Let's catch up later.", 
      timestamp: '11:15 AM',
      contact: {
        name: 'Bob',
        status: 'Last seen today at 10:45 AM',
        avatar: 'https://ui-avatars.com/api/?name=Bob&background=00b2ff&color=fff&rounded=true&size=128'
      }
    },
  ];

  const initialMessages = {
    1: [
      { id: 101, text: 'Hey, how are you?', timestamp: '10:00 AM', sender: { id: 1, name: 'Alice' } },
      { id: 102, text: 'I am good, thanks!', timestamp: '10:02 AM', sender: { id: 3, name: 'Me' } },
    ],
    2: [
      { id: 201, text: "Let's catch up later.", timestamp: '11:15 AM', sender: { id: 2, name: 'Bob' } },
    ],
  };

  const [selectedConversation, setSelectedConversation] = useState(null);
  const [messages, setMessages] = useState([]);
  const currentUser = { id: 3, name: 'Me' };

  const handleSelectConversation = (conversation) => {
    setSelectedConversation(conversation);
    const convoMessages = initialMessages[conversation.id] || [];
    setMessages(convoMessages);
  };

  const handleSendMessage = (text) => {
    if (!selectedConversation) return;
    const newMessage = {
      id: new Date().getTime(),
      text,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      sender: currentUser,
    };
    setMessages(prevMessages => [...prevMessages, newMessage]);
  };

  // Auto-select first conversation if none selected
  useEffect(() => {
    if (!selectedConversation && defaultConversations.length > 0) {
      handleSelectConversation(defaultConversations[0]);
    }
  }, []);

  return (
    <div className="messaging-container">
      <div className="sidebar">
        <ConversationsList
          conversations={defaultConversations}
          onSelectConversation={handleSelectConversation}
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
              <ChatWindow messages={messages} currentUser={currentUser} />
              
              {/* Typing indicator - shown conditionally */}
              {Math.random() > 0.7 && (
                <div className="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              )}
            </div>
            <MessageInput onSendMessage={handleSendMessage} />
          </>
        ) : (
          <p>Please select a conversation</p>
        )}
      </div>
    </div>
  );
};

export default Messaging; 