import api from './api';
import { saveToken, saveUserData, clearAuthData } from '../utils/tokenStorage';

/**
 * Register a new user
 * @param {Object} userData - User registration data
 * @param {string} userData.username - Username
 * @param {string} userData.email - Email
 * @param {string} userData.password - Password
 * @returns {Promise} Promise with registration response
 */
export const register = async (userData) => {
  try {
    const response = await api.post('/auth/register', userData);
    return response.data;
  } catch (error) {
    throw error.response?.data || {
      error: 'Registration failed. Please try again.'
    };
  }
};

/**
 * Login a user
 * @param {Object} credentials - Login credentials
 * @param {string} credentials.email - Email
 * @param {string} credentials.password - Password
 * @returns {Promise} Promise with login response including token and user data
 */
export const login = async (credentials) => {
  try {
    const response = await api.post('/auth/login', credentials);
    const { token, user } = response.data;
    
    // Save token and user data to local storage
    saveToken(token);
    saveUserData(user);
    
    return { token, user };
  } catch (error) {
    throw error.response?.data || {
      error: 'Login failed. Please check your credentials.'
    };
  }
};

/**
 * Logout the current user
 */
export const logout = () => {
  clearAuthData();
};

/**
 * Get the current user profile
 * @returns {Promise} Promise with user profile data
 */
export const getCurrentUser = async () => {
  try {
    const response = await api.get('/auth/me');
    return response.data;
  } catch (error) {
    throw error.response?.data || {
      error: 'Failed to get user profile.'
    };
  }
}; 