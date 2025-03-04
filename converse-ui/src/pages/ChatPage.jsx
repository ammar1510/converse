import { useAuth } from '../context/AuthContext';

/**
 * ChatPage component
 * Placeholder for the chat interface (will be expanded in Phase 2)
 */
const ChatPage = () => {
  const { user } = useAuth();

  return (
    <div className="chat-page">
      <div className="chat-container">
        <div className="chat-header">
          <h2>Welcome to Converse, {user?.username}!</h2>
          <p>Chat functionality coming soon in Phase 2.</p>
        </div>
        <div className="chat-placeholder">
          <p>The chat interface will be implemented here.</p>
          <p>Features will include:</p>
          <ul>
            <li>Real-time messaging with WebSockets</li>
            <li>Message history and loading</li>
            <li>Typing indicators</li>
            <li>Online presence</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ChatPage; 