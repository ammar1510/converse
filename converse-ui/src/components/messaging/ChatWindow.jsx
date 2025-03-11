import React, { useEffect, useRef } from 'react';

const ChatWindow = ({ messages, currentUser }) => {
  const messagesEndRef = useRef(null);
  
  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);
  
  // Format timestamp from ISO string
  const formatTimestamp = (timestamp) => {
    if (!timestamp) return '';
    
    try {
      return new Date(timestamp).toLocaleTimeString([], { 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    } catch (error) {
      return timestamp;
    }
  };

  return (
    <div className="chat-window">
      {messages && messages.length ? (
        <>
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={msg.sender_id === currentUser?.id ? 'message outgoing' : 'message incoming'}
            >
              <div className="sender">
                {msg.sender_id === currentUser?.id ? 'You' : (msg.sender?.username || 'User')}
              </div>
              <div className="text">{msg.content}</div>
              <div className="timestamp">{formatTimestamp(msg.created_at)}</div>
              {!msg.is_read && msg.sender_id !== currentUser?.id && (
                <div className="unread-indicator"></div>
              )}
            </div>
          ))}
          <div ref={messagesEndRef} />
        </>
      ) : (
        <p style={{ textAlign: 'center', color: '#888' }}>No messages in this conversation.</p>
      )}
    </div>
  );
};

export default ChatWindow; 