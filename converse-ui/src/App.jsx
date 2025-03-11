import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ChatProvider } from './context/ChatContext';
import ProtectedRoute from './routes/ProtectedRoute';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import ChatPage from './pages/ChatPage';
import Navbar from './components/layout/Navbar';
import './App.css';

// Debug log
console.log('App component rendering');

/**
 * Main App component
 * Sets up routing and authentication context
 */
function App() {
  console.log('Inside App component function');
  
  // Simplified component for debugging
  try {
    return (
      <AuthProvider>
        <ChatProvider>
          <BrowserRouter>
            <div className="app-container">
              <Navbar />
              <main className="main-content">
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
          </BrowserRouter>
        </ChatProvider>
      </AuthProvider>
    );
  } catch (error) {
    console.error('Error rendering App:', error);
    return <div>Something went wrong. Check console for details.</div>;
  }
}

export default App;
