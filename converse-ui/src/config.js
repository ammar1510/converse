// Configuration for the Converse application

// API URLs
export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

// Auth token settings
export const AUTH_TOKEN_KEY = 'converse_auth_token';
export const USER_DATA_KEY = 'converse_user_data';

// WebSocket settings (for future use)
export const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'; 