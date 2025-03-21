import React, { useState, useEffect } from 'react';

const MessageInput = ({ onSendMessage, onTyping }) => {
  const [message, setMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);

  // Cleanup when component unmounts
  useEffect(() => {
    return () => {
      // Ensure typing indicator is turned off when component unmounts
      if (isTyping && onTyping) {
        onTyping(false);
      }
    };
  }, [onTyping, isTyping]);

  const handleChange = (e) => {
    const newMessage = e.target.value;
    setMessage(newMessage);
    
    // Simple typing indicator logic - just notify parent component
    if (newMessage.trim() !== '' && !isTyping && onTyping) {
      setIsTyping(true);
      onTyping(true);
    } else if (newMessage.trim() === '' && isTyping && onTyping) {
      setIsTyping(false);
      onTyping(false);
    }
  };

  const handleSend = (e) => {
    e.preventDefault();
    if (message.trim() === '') {
      return;
    }
    
    onSendMessage(message);
    setMessage('');
    
    // Turn off typing indicator when message is sent
    if (onTyping && isTyping) {
      setIsTyping(false);
      onTyping(false);
    }
  };

  return (
    <div className="message-input-container">
      <form 
        className="message-input" 
        onSubmit={handleSend}
      >
        <input
          type="text"
          placeholder="Type your message..."
          value={message}
          onChange={handleChange}
        />
        <button type="submit">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M22 2L11 13" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M22 2L15 22L11 13L2 9L22 2Z" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </form>
    </div>
  );
};

export default MessageInput; 