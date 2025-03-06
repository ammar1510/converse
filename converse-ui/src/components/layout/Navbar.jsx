import { Link } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

/**
 * Navbar component
 * Displays navigation links based on authentication status
 */
const Navbar = () => {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <Link to="/" className="navbar-logo">
          Converse
        </Link>
      </div>
      
      <div className="navbar-menu">
        {isAuthenticated ? (
          // Links for authenticated users
          <>
            <span className="navbar-user">Welcome, {user?.username}</span>
            <Link to="/chat" className="navbar-item">
              Chat
            </Link>
            <button 
              className="navbar-button" 
              onClick={logout}
            >
              Logout
            </button>
          </>
        ) : (
          // Links for unauthenticated users
          <>
            <Link to="/login" className="navbar-item">
              Login
            </Link>
            <Link to="/register" className="navbar-item">
              Register
            </Link>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar; 