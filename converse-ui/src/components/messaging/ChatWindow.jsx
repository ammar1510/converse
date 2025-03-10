import React from 'react';

const ChatWindow = ({ messages, currentUser }) => {
  return (
    <div className="chat-window">
      {messages && messages.length ? (
        <>
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={msg.sender.id === currentUser.id ? 'message outgoing' : 'message incoming'}
            >
              <div className="sender">{msg.sender.name}</div>
              <div className="text">{msg.text}</div>
              <div className="timestamp">{msg.timestamp}</div>
            </div>
          ))}
        </>
      ) : (
        <p style={{ textAlign: 'center', color: '#888' }}>No messages in this conversation.</p>
      )}
    </div>
  );
};

export default ChatWindow; 