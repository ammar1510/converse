/**
 * Utility functions for formatting data in the UI
 */

/**
 * Format timestamp from ISO string to localized time
 * @param {string} timestamp - ISO timestamp string
 * @param {Object} options - Formatting options
 * @returns {string} Formatted timestamp
 */
export const formatTimestamp = (timestamp, options = {}) => {
  if (!timestamp) return '';
  
  try {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit',
      ...options
    });
  } catch (error) {
    return timestamp;
  }
};

/**
 * Generate a fallback avatar URL if user has no avatar
 * @param {string} name - User name or identifier
 * @param {Object} options - Avatar options
 * @returns {string} Avatar URL
 */
export const generateAvatarUrl = (name, options = {}) => {
  const defaultOptions = {
    background: '4ead7c',
    color: 'fff',
    rounded: true,
    size: 128
  };
  
  const mergedOptions = { ...defaultOptions, ...options };
  const { background, color, rounded, size } = mergedOptions;
  
  return `https://ui-avatars.com/api/?name=${encodeURIComponent(name)}&background=${background}&color=${color}&rounded=${rounded}&size=${size}`;
}; 