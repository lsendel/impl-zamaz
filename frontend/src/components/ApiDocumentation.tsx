import React from 'react';

interface ApiDocumentationProps {
  token: string | null;
}

const ApiDocumentation: React.FC<ApiDocumentationProps> = ({ token }) => {
  const handleOpenSwagger = () => {
    window.open('http://localhost:8080/swagger/index.html', '_blank');
  };

  return (
    <div className="card">
      <h2>📚 API Documentation</h2>
      <div className="api-docs">
        <p>Interactive API documentation with authentication support:</p>
        
        <button 
          className="swagger-btn"
          onClick={handleOpenSwagger}
          aria-label="Open Swagger API documentation in new tab"
        >
          Open Swagger UI
        </button>
        
        <div className="api-info">
          <h4>Authentication:</h4>
          <ol>
            <li>Click "Authorize" button in Swagger UI</li>
            <li>Enter: Bearer {token ? `${token.substring(0, 20)}...` : '<your-token>'}</li>
            <li>Click "Authorize" to set the token</li>
            <li>Try out the authenticated endpoints!</li>
          </ol>
          
          <div className="token-info">
            <h4>Current Token:</h4>
            <div className="token-display">
              {token ? (
                <>
                  <code>{token.substring(0, 30)}...</code>
                  <button 
                    className="copy-token-btn"
                    onClick={() => navigator.clipboard.writeText(`Bearer ${token}`)}
                    aria-label="Copy token to clipboard"
                    title="Copy token to clipboard"
                  >
                    📋
                  </button>
                </>
              ) : (
                <em>No token available</em>
              )}
            </div>
          </div>
          
          <div className="api-endpoints">
            <h4>Available Endpoints:</h4>
            <ul>
              <li><strong>GET /health</strong> - Health check (public)</li>
              <li><strong>POST /api/v1/auth/login</strong> - User login (public)</li>
              <li><strong>GET /api/v1/trust-score</strong> - Get trust score (authenticated)</li>
              <li><strong>GET /api/v1/user/profile</strong> - User profile (authenticated)</li>
              <li><strong>GET /api/v1/protected</strong> - Protected resource (authenticated)</li>
              <li><strong>GET /api/v1/discovery/services</strong> - Service discovery (public)</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ApiDocumentation;