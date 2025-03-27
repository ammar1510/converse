import { createContext, useState, useEffect, useContext } from 'react';
import { isTokenValid, getUserData, clearAuthData } from '../utils/tokenStorage';
import * as authService from '../services/authService';
import websocketService from '../services/websocketService';
import messageService from '../services/messageService';

// Create context
const AuthContext = createContext(null);

/**
 * AuthProvider component to wrap the application with authentication context
 */
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Check for existing auth on initial load
  useEffect(() => {
    const initAuth = async () => {
      try {
        if (isTokenValid()) {
          // Get user from local storage
          const userData = getUserData();
          if (userData) {
            setUser(userData);
            setIsAuthenticated(true);
          } else {
            // If token is valid but no user data, try to fetch it
            try {
              const currentUser = await authService.getCurrentUser();
              setUser(currentUser);
              setIsAuthenticated(true);
            } catch (e) {
              // If we can't get the current user, clear auth data
              clearAuthData();
              setIsAuthenticated(false);
            }
          }
        } else {
          // Token is invalid or doesn't exist
          clearAuthData();
          setIsAuthenticated(false);
        }
      } catch (err) {
        console.error('Auth initialization error:', err);
        setError('Authentication failed. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    initAuth();
  }, []);

  /**
   * Login a user with email and password
   */
  const login = async (email, password) => {
    setLoading(true);
    setError(null);
    
    try {
      // First, ensure we're logged out completely
      authService.logout();
      
      // Reset WebSocket connection state to ensure a clean start
      websocketService.reset();
      
      // Now perform the login
      const { user } = await authService.login({ email, password });
      
      // Update state after successful login
      setUser(user);
      setIsAuthenticated(true);
      
      return user;
    } catch (err) {
      setError(err.error || 'Login failed. Please check your credentials.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  /**
   * Register a new user
   */
  const register = async (username, email, password) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await authService.register({ username, email, password });
      return result;
    } catch (err) {
      setError(err.error || 'Registration failed. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  /**
   * Logout the current user
   */
  const logout = async () => {
    try {
      // Reset WebSocket connection state before clearing auth data
      websocketService.reset();
      
      // Clear cached API data
      messageService.invalidateCache();
      
      // Perform logout
      authService.logout();
      
      // Update state after logout
      setUser(null);
      setIsAuthenticated(false);
      
      console.log('User logged out, WebSocket reset, auth data cleared, cache invalidated');
    } catch (err) {
      console.error('Error during logout:', err);
      setError('Logout failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Context value
  const value = {
    user,
    isAuthenticated,
    loading,
    error,
    login,
    register,
    logout
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

/**
 * Custom hook to use the auth context
 */
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}; 