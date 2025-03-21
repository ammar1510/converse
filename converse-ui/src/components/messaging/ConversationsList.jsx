import React, { useState } from 'react';
import { generateAvatarUrl } from '../../utils/formatUtils';

const ConversationsList = ({ conversations, onSelectConversation }) => {
  const [activeConversation, setActiveConversation] = useState(null);
  
  const defaultConversations = conversations || [
    { id: 1, title: 'Conversation 1', lastMessage: 'Hey, how are you?', timestamp: '10:00 AM' },
    { id: 2, title: 'Conversation 2', lastMessage: "Let's catch up later.", timestamp: '11:15 AM' },
  ];

  const handleSelectConversation = (convo) => {
    setActiveConversation(convo.id);
    onSelectConversation(convo);
  };

  return (
    <div className="conversations-list">
      <h3>Conversations</h3>
      <ul>
        {defaultConversations.map(convo => (
          <li 
            key={convo.id} 
            className={`conversation-item ${activeConversation === convo.id ? 'active' : ''}`}
            onClick={() => handleSelectConversation(convo)}
          >
            {convo.contact && (
              <div className="conversation-avatar">
                <img 
                  src={convo.contact.avatar || generateAvatarUrl(convo.contact.name)} 
                  alt={convo.contact.name} 
                />
              </div>
            )}
            <div className="conversation-content">
              <div className="conversation-title">{convo.contact ? convo.contact.name : convo.title}</div>
              <div className="conversation-preview">{convo.lastMessage}</div>
            </div>
            <div className="conversation-time">{convo.timestamp}</div>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ConversationsList; 