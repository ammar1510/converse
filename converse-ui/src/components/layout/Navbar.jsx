import { Link } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useState } from 'react';

/**
 * Navbar component
 * Displays navigation links based on authentication status
 */
const Navbar = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);

  const handleLogoutClick = () => {
    setShowLogoutConfirm(true);
  };

  const handleConfirmLogout = () => {
    logout();
    setShowLogoutConfirm(false);
  };

  const handleCancelLogout = () => {
    setShowLogoutConfirm(false);
  };

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
              onClick={handleLogoutClick}
            >
              Logout
            </button>
            
            {/* Logout confirmation dialog */}
            {showLogoutConfirm && (
              <div className="logout-confirm-overlay">
                <div className="logout-confirm-dialog">
                  <h3>Are you sure?</h3>
                  <p>You're about to log out of your account.</p>
                  <div className="logout-confirm-actions">
                    <button 
                      className="logout-cancel-btn" 
                      onClick={handleCancelLogout}
                    >
                      Cancel
                    </button>
                    <button 
                      className="logout-confirm-btn" 
                      onClick={handleConfirmLogout}
                    >
                      Yes, Logout
                    </button>
                  </div>
                </div>
              </div>
            )}
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