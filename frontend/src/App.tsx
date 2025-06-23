import React, { useState, useEffect } from 'react';
import './App.css';

// Components
import ErrorBoundary from './ErrorBoundary';
import ErrorDisplay from './components/ErrorDisplay';
import LoginForm from './components/LoginForm';
import UserInfo from './components/UserInfo';
import TrustScore from './components/TrustScore';
import ServiceDiscovery from './components/ServiceDiscovery';
import ApiDocumentation from './components/ApiDocumentation';

// Hooks
import { useError } from './hooks/useError';

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
  const { error, showError, clearError, handleError, handleApiError } = useError();

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
    clearError();
    
    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        await handleApiError(response);
        return;
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
      handleError(err);
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    setTrustScore(null);
    setServices([]);
    clearError();
    localStorage.removeItem('token');
  };

  const fetchUserInfo = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/user/profile', {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      if (response.ok) {
        const data = await response.json();
        setUser(data);
      } else {
        await handleApiError(response);
      }
    } catch (err) {
      handleError(err);
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
      } else {
        await handleApiError(response);
      }
    } catch (err) {
      handleError(err);
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
    return (
      <ErrorBoundary>
        <LoginForm onLogin={login} loading={loading} error={error?.message || null} />
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <div className="app">
        <header className="app-header">
          <h1>🔐 Zero Trust Dashboard</h1>
          <button onClick={logout} className="logout-btn" aria-label="Logout from application">
            Logout
          </button>
        </header>

        {error && (
          <ErrorDisplay 
            error={error} 
            onDismiss={clearError} 
            variant="banner" 
          />
        )}

        <div className="container">
          <UserInfo user={user} />
          <TrustScore trustScore={trustScore} />
          <ServiceDiscovery services={services} canAccessService={canAccessService} />
          <ApiDocumentation token={token} />
        </div>
      </div>
    </ErrorBoundary>
  );
};

export default App;