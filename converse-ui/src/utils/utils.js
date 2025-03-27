/**
 * General utility functions for the application
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

/**
 * Creates a standardized event handler with cleanup function
 * @param {Object} eventSource - The source object that emits events (e.g., websocketService)
 * @param {string} eventType - The type of event to listen for
 * @param {Function} handler - The event handler function
 * @returns {Function} Cleanup function to remove the event listener
 */
export const createEventHandler = (eventSource, eventType, handler) => {
  if (!eventSource || !eventSource.on || !eventSource.off) {
    console.error('Invalid event source provided to createEventHandler');
    return () => {};
  }
  
  // Add the event listener
  eventSource.on(eventType, handler);
  
  // Return a cleanup function
  return () => {
    eventSource.off(eventType, handler);
  };
};

/**
 * Register multiple event handlers at once and return a single cleanup function
 * @param {Object} eventSource - The source object that emits events
 * @param {Object} handlers - An object mapping event types to handler functions
 * @returns {Function} Cleanup function to remove all event listeners
 */
export const registerEventHandlers = (eventSource, handlers) => {
  if (!eventSource || !handlers || typeof handlers !== 'object') {
    console.error('Invalid parameters provided to registerEventHandlers');
    return () => {};
  }
  
  // Create an array of cleanup functions
  const cleanupFunctions = Object.entries(handlers).map(
    ([eventType, handler]) => createEventHandler(eventSource, eventType, handler)
  );
  
  // Return a single cleanup function that calls all cleanup functions
  return () => {
    cleanupFunctions.forEach(cleanup => cleanup());
  };
}; 