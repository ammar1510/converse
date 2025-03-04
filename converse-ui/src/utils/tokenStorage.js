import { AUTH_TOKEN_KEY, USER_DATA_KEY } from '../config';
import { jwtDecode } from 'jwt-decode';

/**
 * Save authentication token to local storage
 * @param {string} token - JWT token
 */
export const saveToken = (token) => {
  localStorage.setItem(AUTH_TOKEN_KEY, token);
};

/**
 * Get authentication token from local storage
 * @returns {string|null} Token or null if not found
 */
export const getToken = () => {
  return localStorage.getItem(AUTH_TOKEN_KEY);
};

/**
 * Remove authentication token from local storage
 */
export const removeToken = () => {
  localStorage.removeItem(AUTH_TOKEN_KEY);
};

/**
 * Check if token exists and is valid (not expired)
 * @returns {boolean} True if token is valid
 */
export const isTokenValid = () => {
  const token = getToken();
  if (!token) return false;

  try {
    const decoded = jwtDecode(token);
    const currentTime = Date.now() / 1000;
    
    // Check if token is expired
    return decoded.exp > currentTime;
  } catch (error) {
    console.error('Error decoding token:', error);
    return false;
  }
};

/**
 * Save user data to local storage
 * @param {Object} userData - User data object
 */
export const saveUserData = (userData) => {
  localStorage.setItem(USER_DATA_KEY, JSON.stringify(userData));
};

/**
 * Get user data from local storage
 * @returns {Object|null} User data or null if not found
 */
export const getUserData = () => {
  const data = localStorage.getItem(USER_DATA_KEY);
  return data ? JSON.parse(data) : null;
};

/**
 * Remove user data from local storage
 */
export const removeUserData = () => {
  localStorage.removeItem(USER_DATA_KEY);
};

/**
 * Clear all authentication data (both token and user data)
 */
export const clearAuthData = () => {
  removeToken();
  removeUserData();
}; 