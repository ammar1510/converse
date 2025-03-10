import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

/**
 * LoginForm component
 * Handles user login with improved UI
 */
const LoginForm = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const { login, error } = useAuth();
  const navigate = useNavigate();

  /**
   * Handle form submission
   */
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    if (!email || !password) {
      setFormError('Please fill in all fields');
      return;
    }
    
    setFormError('');
    setIsSubmitting(true);
    
    try {
      await login(email, password);
      // Redirect to chat page on successful login
      navigate('/chat');
    } catch (err) {
      setFormError(err.error || 'Login failed. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="auth-form-container">
      <h2>Welcome Back</h2>
      
      {(formError || error) && (
        <div className="auth-error">
          {formError || error}
        </div>
      )}
      
      <form className="auth-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            type="email"
            id="email"
            placeholder="Enter your email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={isSubmitting}
            required
          />
        </div>
        
        <button 
          type="submit" 
          className="auth-button"
          disabled={isSubmitting}
        >
          {isSubmitting ? 'LOGGING IN...' : 'LOGIN'}
        </button>
      </form>
      
      <div className="auth-links">
        <p>
          Don't have an account? <Link to="/register">Register Here</Link>
        </p>
      </div>
    </div>
  );
};

export default LoginForm; 