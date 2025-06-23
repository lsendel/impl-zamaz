import React from 'react';

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

interface TrustScoreProps {
  trustScore: TrustScore | null;
}

interface FactorDisplayProps {
  label: string;
  icon: string;
  value: number;
  maxValue: number;
}

const FactorDisplay: React.FC<FactorDisplayProps> = ({ label, icon, value, maxValue }) => (
  <div className="factor">
    <span>{icon} {label}</span>
    <div className="progress" role="progressbar" aria-valuenow={value} aria-valuemax={maxValue}>
      <div 
        className="progress-bar" 
        style={{ width: `${(value / maxValue) * 100}%` }}
        aria-label={`${label}: ${value} out of ${maxValue}`}
      ></div>
    </div>
    <span>{value}/{maxValue}</span>
  </div>
);

const TrustScore: React.FC<TrustScoreProps> = ({ trustScore }) => {
  if (!trustScore) {
    return (
      <div className="card">
        <h2>📊 Trust Score</h2>
        <p>Loading trust score...</p>
      </div>
    );
  }

  const getScoreColor = (score: number): string => {
    if (score >= 80) return '#4CAF50';
    if (score >= 60) return '#FF9800';
    if (score >= 40) return '#FF5722';
    return '#F44336';
  };

  const factors = [
    { key: 'identity', label: 'Identity', icon: '🔐', maxValue: 30 },
    { key: 'device', label: 'Device', icon: '📱', maxValue: 25 },
    { key: 'behavior', label: 'Behavior', icon: '🔍', maxValue: 20 },
    { key: 'location', label: 'Location', icon: '🌍', maxValue: 15 },
    { key: 'risk', label: 'Risk', icon: '⚠️', maxValue: 10 },
  ] as const;

  return (
    <div className="card">
      <h2>📊 Trust Score</h2>
      <div className="trust-score">
        <div className="overall-score">
          <div 
            className="score-circle" 
            style={{ 
              background: `conic-gradient(${getScoreColor(trustScore.overall)} ${trustScore.overall * 3.6}deg, #e0e0e0 0deg)` 
            }}
            role="img"
            aria-label={`Overall trust score: ${trustScore.overall} out of 100`}
          >
            <span>{trustScore.overall}</span>
          </div>
        </div>
        
        <div className="factors">
          <h3>Trust Factors:</h3>
          {factors.map(factor => (
            <FactorDisplay
              key={factor.key}
              label={factor.label}
              icon={factor.icon}
              value={trustScore.factors[factor.key]}
              maxValue={factor.maxValue}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

export default TrustScore;