import React, { useState } from 'react';

interface LoginFormProps {
  onLogin: (username: string, password: string) => void;
  loading: boolean;
  error: string | null;
}

const LoginForm: React.FC<LoginFormProps> = ({ onLogin, loading, error }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (username.trim() && password.trim()) {
      onLogin(username.trim(), password);
    }
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1>🔐 Zero Trust Login</h1>
        
        <form onSubmit={handleSubmit} noValidate>
          <div className="form-group">
            <label htmlFor="username" className="form-label">
              Username
            </label>
            <input
              id="username"
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={loading}
              autoComplete="username"
              aria-describedby={error ? "error-message" : undefined}
              className={error ? "form-input error" : "form-input"}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password" className="form-label">
              Password
            </label>
            <div className="password-input-container">
              <input
                id="password"
                type={showPassword ? "text" : "password"}
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={loading}
                autoComplete="current-password"
                aria-describedby={error ? "error-message" : undefined}
                className={error ? "form-input error" : "form-input"}
              />
              <button
                type="button"
                onClick={togglePasswordVisibility}
                disabled={loading}
                className="password-toggle"
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
          </div>

          {error && (
            <div 
              id="error-message" 
              className="error" 
              role="alert"
              aria-live="polite"
            >
              {error}
            </div>
          )}

          <button 
            type="submit" 
            disabled={loading || !username.trim() || !password.trim()}
            className="login-button"
            aria-describedby="login-help"
          >
            {loading ? (
              <>
                <span className="loading-spinner" aria-hidden="true">⏳</span>
                Logging in...
              </>
            ) : (
              'Login'
            )}
          </button>
        </form>

        <div className="demo-info">
          <h4>Demo Credentials:</h4>
          <div className="credential-options">
            <div className="credential-option">
              <strong>Admin:</strong> admin / admin
            </div>
            <div className="credential-option">
              <strong>User:</strong> user / user
            </div>
            <div className="credential-option">
              <strong>Test:</strong> test / test
            </div>
          </div>
          
          <div className="quick-login">
            <button
              type="button"
              onClick={() => {
                setUsername('admin');
                setPassword('admin');
              }}
              disabled={loading}
              className="quick-login-btn"
            >
              Use Admin Credentials
            </button>
          </div>
        </div>

        <div className="security-info">
          <h4>🛡️ Zero Trust Security Features:</h4>
          <ul>
            <li>Multi-factor authentication ready</li>
            <li>Device attestation verification</li>
            <li>Risk-based access control</li>
            <li>Continuous trust evaluation</li>
            <li>JWT token-based authentication</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default LoginForm;