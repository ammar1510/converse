import React, { useState, useEffect, useRef } from 'react';

const MessageInput = ({ onSendMessage, onTyping }) => {
  const [message, setMessage] = useState('');
  const typingTimeoutRef = useRef(null);
  const isTypingRef = useRef(false);

  // Handle typing indicator
  useEffect(() => {
    return () => {
      // Clean up typing timeout on unmount
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      
      // Ensure typing indicator is turned off when component unmounts
      if (isTypingRef.current && onTyping) {
        onTyping(false);
      }
    };
  }, [onTyping]);

  const handleChange = (e) => {
    const newMessage = e.target.value;
    setMessage(newMessage);
    
    // Handle typing indicator
    if (onTyping) {
      // If user wasn't typing before, send typing indicator
      if (!isTypingRef.current) {
        isTypingRef.current = true;
        onTyping(true);
      }
      
      // Clear existing timeout
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      
      // Set timeout to stop typing indicator after 2 seconds of inactivity
      typingTimeoutRef.current = setTimeout(() => {
        isTypingRef.current = false;
        onTyping(false);
      }, 2000);
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
    if (onTyping && isTypingRef.current) {
      isTypingRef.current = false;
      onTyping(false);
      
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
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