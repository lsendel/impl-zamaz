import React, { useState, useEffect } from 'react';
import './App.css';

interface User {
  id: string;
  username: string;
  email: string;
  roles: string[];
}

interface TrustScore {
  overall: number;
  factors: {
    identity: number;
    device: number;
    behavior: number;
    location: number;
    risk: number;
  };
}

interface Service {
  name: string;
  url: string;
  status: 'healthy' | 'unhealthy' | 'unknown';
  trustRequired: number;
}

const App: React.FC = () => {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [user, setUser] = useState<User | null>(null);
  const [trustScore, setTrustScore] = useState<TrustScore | null>(null);
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Service discovery
  const discoveredServices: Service[] = [
    { name: 'API Gateway', url: 'http://localhost:8080', status: 'unknown', trustRequired: 0 },
    { name: 'Keycloak', url: 'http://localhost:8082', status: 'unknown', trustRequired: 0 },
    { name: 'User Service', url: 'http://localhost:8081', status: 'unknown', trustRequired: 25 },
    { name: 'Admin Service', url: 'http://localhost:8083', status: 'unknown', trustRequired: 75 },
    { name: 'Audit Service', url: 'http://localhost:8084', status: 'unknown', trustRequired: 50 },
  ];

  useEffect(() => {
    if (token) {
      fetchUserInfo();
      fetchTrustScore();
      checkServices();
    }
  }, [token]);

  const login = async (username: string, password: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        throw new Error('Login failed');
      }

      const data = await response.json();
      setToken(data.access_token);
      localStorage.setItem('token', data.access_token);
      setUser(data.user);
      setTrustScore({
        overall: data.trust_score,
        factors: {
          identity: 30,
          device: 20,
          behavior: 18,
          location: 12,
          risk: 8,
        },
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    setTrustScore(null);
    localStorage.removeItem('token');
  };

  const fetchUserInfo = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/user/me', {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      if (response.ok) {
        const data = await response.json();
        setUser(data);
      }
    } catch (err) {
      console.error('Failed to fetch user info:', err);
    }
  };

  const fetchTrustScore = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/trust-score', {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      if (response.ok) {
        const data = await response.json();
        setTrustScore({
          overall: data.overall,
          factors: data.factors,
        });
      }
    } catch (err) {
      console.error('Failed to fetch trust score:', err);
    }
  };

  const checkServices = async () => {
    const updatedServices = await Promise.all(
      discoveredServices.map(async (service) => {
        try {
          const response = await fetch(`${service.url}/health`);
          return {
            ...service,
            status: response.ok ? 'healthy' : 'unhealthy',
          };
        } catch {
          return { ...service, status: 'unhealthy' };
        }
      })
    );
    setServices(updatedServices);
  };

  const canAccessService = (requiredTrust: number): boolean => {
    return trustScore ? trustScore.overall >= requiredTrust : false;
  };

  if (!token) {
    return <LoginForm onLogin={login} loading={loading} error={error} />;
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>🔐 Zero Trust Dashboard</h1>
        <button onClick={logout} className="logout-btn">Logout</button>
      </header>

      <div className="container">
        {/* User Info */}
        <div className="card">
          <h2>👤 User Information</h2>
          {user && (
            <div className="user-info">
              <p><strong>Username:</strong> {user.username}</p>
              <p><strong>Email:</strong> {user.email}</p>
              <p><strong>Roles:</strong> {user.roles.join(', ')}</p>
            </div>
          )}
        </div>

        {/* Trust Score */}
        <div className="card">
          <h2>📊 Trust Score</h2>
          {trustScore && (
            <div className="trust-score">
              <div className="overall-score">
                <div className="score-circle" style={{ 
                  background: `conic-gradient(#4CAF50 ${trustScore.overall * 3.6}deg, #e0e0e0 0deg)` 
                }}>
                  <span>{trustScore.overall}</span>
                </div>
              </div>
              <div className="factors">
                <h3>Trust Factors:</h3>
                <div className="factor">
                  <span>🔐 Identity</span>
                  <div className="progress">
                    <div className="progress-bar" style={{ width: `${(trustScore.factors.identity / 30) * 100}%` }}></div>
                  </div>
                  <span>{trustScore.factors.identity}/30</span>
                </div>
                <div className="factor">
                  <span>📱 Device</span>
                  <div className="progress">
                    <div className="progress-bar" style={{ width: `${(trustScore.factors.device / 25) * 100}%` }}></div>
                  </div>
                  <span>{trustScore.factors.device}/25</span>
                </div>
                <div className="factor">
                  <span>🔍 Behavior</span>
                  <div className="progress">
                    <div className="progress-bar" style={{ width: `${(trustScore.factors.behavior / 20) * 100}%` }}></div>
                  </div>
                  <span>{trustScore.factors.behavior}/20</span>
                </div>
                <div className="factor">
                  <span>🌍 Location</span>
                  <div className="progress">
                    <div className="progress-bar" style={{ width: `${(trustScore.factors.location / 15) * 100}%` }}></div>
                  </div>
                  <span>{trustScore.factors.location}/15</span>
                </div>
                <div className="factor">
                  <span>⚠️ Risk</span>
                  <div className="progress">
                    <div className="progress-bar" style={{ width: `${(trustScore.factors.risk / 10) * 100}%` }}></div>
                  </div>
                  <span>{trustScore.factors.risk}/10</span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Service Discovery */}
        <div className="card">
          <h2>🔍 Service Discovery</h2>
          <div className="services">
            {services.map((service) => (
              <div key={service.name} className={`service ${!canAccessService(service.trustRequired) ? 'locked' : ''}`}>
                <div className="service-header">
                  <h3>{service.name}</h3>
                  <span className={`status ${service.status}`}>
                    {service.status === 'healthy' ? '✅' : '❌'} {service.status}
                  </span>
                </div>
                <p className="service-url">{service.url}</p>
                <p className="trust-requirement">
                  Required Trust: {service.trustRequired}
                  {!canAccessService(service.trustRequired) && (
                    <span className="access-denied"> 🔒 Access Denied</span>
                  )}
                </p>
                {canAccessService(service.trustRequired) && (
                  <button 
                    className="access-btn"
                    onClick={() => window.open(`${service.url}/swagger/index.html`, '_blank')}
                  >
                    Access Service
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* API Documentation */}
        <div className="card">
          <h2>📚 API Documentation</h2>
          <div className="api-docs">
            <p>Interactive API documentation with authentication support:</p>
            <button 
              className="swagger-btn"
              onClick={() => window.open('http://localhost:8080/swagger/index.html', '_blank')}
            >
              Open Swagger UI
            </button>
            <div className="api-info">
              <h4>Authentication:</h4>
              <p>1. Click "Authorize" button in Swagger UI</p>
              <p>2. Enter: Bearer {token ? `${token.substring(0, 20)}...` : '<your-token>'}</p>
              <p>3. Click "Authorize" to set the token</p>
              <p>4. Try out the authenticated endpoints!</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Login Form Component
const LoginForm: React.FC<{
  onLogin: (username: string, password: string) => void;
  loading: boolean;
  error: string | null;
}> = ({ onLogin, loading, error }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onLogin(username, password);
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1>🔐 Zero Trust Login</h1>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          {error && <p className="error">{error}</p>}
          <button type="submit" disabled={loading}>
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>
        <div className="demo-info">
          <p>Demo credentials:</p>
          <p>Username: admin | Password: admin</p>
        </div>
      </div>
    </div>
  );
};

export default App;