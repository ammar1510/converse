import { useAuth } from '../context/AuthContext';
import Messaging from '../components/messaging/Messaging';

/**
 * ChatPage component
 * Implements the chat interface using the Messaging component
 */
const ChatPage = () => {
  const { user } = useAuth();

  return (
    <div className="chat-page">
      <div className="chat-container">
        <div className="chat-header">
          <div className="welcome-message">
            <span className="wave-emoji">👋</span>
            <h2>Hey <span className="username-highlight">{user?.username}</span>! Ready to chat?</h2>
          </div>
        </div>
        <div className="messaging-wrapper">
          <Messaging />
        </div>
      </div>
    </div>
  );
};

export default ChatPage; 