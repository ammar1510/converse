import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ChatProvider } from './context/ChatContext';
import ProtectedRoute from './routes/ProtectedRoute';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import ChatPage from './pages/ChatPage';
import Navbar from './components/layout/Navbar';
import './App.css';

/**
 * Layout component that conditionally renders Navbar
 */
function AppLayout() {
  const location = useLocation();
  const isChatPage = location.pathname === '/chat';
  
  return (
    <div className="app-container">
      <Navbar />
      <main className={`main-content ${isChatPage ? 'full-height' : ''}`}>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          
          {/* Protected routes */}
          <Route element={<ProtectedRoute />}>
            <Route path="/chat" element={<ChatPage />} />
          </Route>
          
          {/* Redirects */}
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </main>
    </div>
  );
}

/**
 * Main App component
 * Sets up routing and authentication context
 */
function App() {
  return (
    <AuthProvider>
      <ChatProvider>
        <BrowserRouter>
          <AppLayout />
        </BrowserRouter>
      </ChatProvider>
    </AuthProvider>
  );
}

export default App;
