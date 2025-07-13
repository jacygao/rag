import React from 'react';
import { Reference } from '../types';
import './References.css';

interface ReferencesProps {
  references: Reference[];
}

const References: React.FC<ReferencesProps> = ({ references }) => {
  const getSourceIcon = (source: string) => {
    switch (source.toLowerCase()) {
      case 'confluence':
        return '📝';
      case 'slack':
        return '💬';
      case 'gmail':
        return '📧';
      default:
        return '🔗';
    }
  };

  const getSourceColor = (source: string) => {
    switch (source.toLowerCase()) {
      case 'confluence':
        return '#0052cc';
      case 'slack':
        return '#4a154b';
      case 'gmail':
        return '#ea4335';
      default:
        return '#666';
    }
  };

  return (
    <div className="references">
      <div className="references-header">
        <span className="references-title">Sources</span>
        <span className="references-count">{references.length}</span>
      </div>
      
      <div className="references-list">
        {references.map((ref, index) => (
          <a
            key={index}
            href={ref.url}
            target="_blank"
            rel="noopener noreferrer"
            className="reference-item"
          >
            <div className="reference-icon">
              {getSourceIcon(ref.source)}
            </div>
            
            <div className="reference-content">
              <div className="reference-title">{ref.title}</div>
              <div 
                className="reference-source"
                style={{ color: getSourceColor(ref.source) }}
              >
                {ref.source.charAt(0).toUpperCase() + ref.source.slice(1)}
              </div>
            </div>
            
            <div className="reference-arrow">
              →
            </div>
          </a>
        ))}
      </div>
    </div>
  );
};

export default References;