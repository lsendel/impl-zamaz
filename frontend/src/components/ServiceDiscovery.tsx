import React from 'react';

interface Service {
  name: string;
  url: string;
  status: 'healthy' | 'unhealthy' | 'unknown';
  trustRequired: number;
}

interface ServiceDiscoveryProps {
  services: Service[];
  canAccessService: (requiredTrust: number) => boolean;
}

interface ServiceCardProps {
  service: Service;
  canAccess: boolean;
}

const ServiceCard: React.FC<ServiceCardProps> = ({ service, canAccess }) => {
  const getStatusIcon = (status: Service['status']) => {
    switch (status) {
      case 'healthy': return '✅';
      case 'unhealthy': return '❌';
      case 'unknown': return '❓';
      default: return '❓';
    }
  };

  const handleAccessService = () => {
    if (canAccess) {
      window.open(`${service.url}/swagger/index.html`, '_blank');
    }
  };

  return (
    <div 
      className={`service ${!canAccess ? 'locked' : ''}`}
      role="article"
      aria-label={`Service: ${service.name}`}
    >
      <div className="service-header">
        <h3>{service.name}</h3>
        <span 
          className={`status ${service.status}`}
          role="img"
          aria-label={`Status: ${service.status}`}
        >
          {getStatusIcon(service.status)} {service.status}
        </span>
      </div>
      
      <p className="service-url">{service.url}</p>
      
      <p className="trust-requirement">
        Required Trust: {service.trustRequired}
        {!canAccess && (
          <span className="access-denied" role="alert"> 🔒 Access Denied</span>
        )}
      </p>
      
      {canAccess && (
        <button 
          className="access-btn"
          onClick={handleAccessService}
          aria-label={`Access ${service.name} service`}
        >
          Access Service
        </button>
      )}
    </div>
  );
};

const ServiceDiscovery: React.FC<ServiceDiscoveryProps> = ({ services, canAccessService }) => {
  return (
    <div className="card">
      <h2>🔍 Service Discovery</h2>
      <div className="services" role="list">
        {services.length === 0 ? (
          <p>No services available. Checking service status...</p>
        ) : (
          services.map((service) => (
            <ServiceCard
              key={service.name}
              service={service}
              canAccess={canAccessService(service.trustRequired)}
            />
          ))
        )}
      </div>
    </div>
  );
};

export default ServiceDiscovery;