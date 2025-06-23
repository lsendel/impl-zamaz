import React from 'react';
import { AppError } from '../hooks/useError';

interface ErrorDisplayProps {
  error: AppError;
  onDismiss: () => void;
  variant?: 'banner' | 'modal' | 'inline';
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({ 
  error, 
  onDismiss, 
  variant = 'banner' 
}) => {
  const getErrorIcon = (code?: string) => {
    if (code?.startsWith('AUTH_')) return '🔒';
    if (code?.startsWith('HTTP_4')) return '⚠️';
    if (code?.startsWith('HTTP_5')) return '🚨';
    if (code?.startsWith('NETWORK_')) return '📡';
    return '❌';
  };

  const getErrorTitle = (code?: string) => {
    if (code?.startsWith('AUTH_')) return 'Authentication Error';
    if (code?.startsWith('HTTP_401')) return 'Unauthorized';
    if (code?.startsWith('HTTP_403')) return 'Access Denied';
    if (code?.startsWith('HTTP_404')) return 'Not Found';
    if (code?.startsWith('HTTP_5')) return 'Server Error';
    if (code?.startsWith('NETWORK_')) return 'Network Error';
    return 'Error';
  };

  const baseClassName = `error-display error-display--${variant}`;

  return (
    <div className={baseClassName} role="alert" aria-live="polite">
      <div className="error-content">
        <div className="error-header">
          <span className="error-icon" aria-hidden="true">
            {getErrorIcon(error.code)}
          </span>
          <h3 className="error-title">{getErrorTitle(error.code)}</h3>
          <button 
            className="error-dismiss"
            onClick={onDismiss}
            aria-label="Dismiss error"
          >
            ✕
          </button>
        </div>
        
        <div className="error-body">
          <p className="error-message">{error.message}</p>
          
          {error.code && (
            <p className="error-code">
              <strong>Error Code:</strong> {error.code}
            </p>
          )}
          
          <p className="error-timestamp">
            <strong>Time:</strong> {error.timestamp.toLocaleString()}
          </p>
          
          {process.env.NODE_ENV === 'development' && error.details && (
            <details className="error-details-dev">
              <summary>Technical Details (Development)</summary>
              <pre>{JSON.stringify(error.details, null, 2)}</pre>
            </details>
          )}
        </div>
        
        <div className="error-actions">
          <button 
            className="error-action-primary"
            onClick={onDismiss}
          >
            Dismiss
          </button>
          
          {error.code?.startsWith('NETWORK_') && (
            <button 
              className="error-action-secondary"
              onClick={() => window.location.reload()}
            >
              Retry
            </button>
          )}
          
          {error.code?.startsWith('AUTH_') && (
            <button 
              className="error-action-secondary"
              onClick={() => {
                localStorage.removeItem('token');
                window.location.href = '/login';
              }}
            >
              Re-login
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default ErrorDisplay;