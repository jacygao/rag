import React, { useState } from 'react';
import { ApiKeys } from '../types';
import './Settings.css';

interface SettingsProps {
  apiKeys: ApiKeys;
  onUpdateKeys: (keys: Partial<ApiKeys>) => void;
}

const Settings: React.FC<SettingsProps> = ({ apiKeys, onUpdateKeys }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleKeyChange = (source: keyof ApiKeys, value: string) => {
    onUpdateKeys({ [source]: value });
  };

  const getConnectionStatus = (key: string) => {
    return key ? 'connected' : 'disconnected';
  };

  const handleOAuthConnect = async (source: keyof ApiKeys) => {
    try {
      let authUrl = '';
      let successMessageType = '';

      if (source === 'confluence') {
        const response = await fetch('/api/auth/confluence');
        const data = await response.json();
        authUrl = data.auth_url;
        successMessageType = 'CONFLUENCE_AUTH_SUCCESS';
      } else if (source === 'gmail') {
        const response = await fetch('/api/auth/gmail');
        const data = await response.json();
        authUrl = data.auth_url;
        successMessageType = 'GMAIL_AUTH_SUCCESS';
      } else if (source === 'slack') {
        const response = await fetch('/api/auth/slack');
        const data = await response.json();
        authUrl = data.auth_url;
        successMessageType = 'SLACK_AUTH_SUCCESS';
      }

      if (!authUrl) return;

      // Open popup window
      const popup = window.open(
        authUrl,
        'oauth',
        'width=500,height=600,scrollbars=yes,resizable=yes'
      );

      // Listen for messages from the popup
      const handleMessage = (event: MessageEvent) => {
        // Allow messages from localhost (development)
        if (!event.origin.includes('localhost')) {
          return;
        }
        
        if (event.data.type === successMessageType) {
          onUpdateKeys({ [source]: event.data.token });
          window.removeEventListener('message', handleMessage);
          popup?.close();
        }
      };

      window.addEventListener('message', handleMessage);

      // Also check if popup is closed manually
      const checkClosed = setInterval(() => {
        if (popup?.closed) {
          clearInterval(checkClosed);
          window.removeEventListener('message', handleMessage);
        }
      }, 1000);

      popup?.focus();
    } catch (error) {
      console.error('OAuth error:', error);
    }
  };

  return (
    <div className="settings">
      <div className="settings-header" onClick={() => setIsExpanded(!isExpanded)}>
        <h3>Data Sources</h3>
        <span className={`expand-icon ${isExpanded ? 'expanded' : ''}`}>▼</span>
      </div>
      
      {isExpanded && (
        <div className="settings-content">
          <div className="source-setting">
            <div className="source-header">
              <span className="source-name">Confluence</span>
              <span className={`status ${getConnectionStatus(apiKeys.confluence)}`}>
                {getConnectionStatus(apiKeys.confluence)}
              </span>
            </div>
            {apiKeys.confluence ? (
              <div className="connected-info">
                <span className="connected-text">✓ Connected to Confluence</span>
                <button 
                  onClick={() => handleKeyChange('confluence', '')}
                  className="disconnect-button"
                >
                  Disconnect
                </button>
              </div>
            ) : (
              <button 
                onClick={() => handleOAuthConnect('confluence')}
                className="connect-button"
              >
                Connect to Confluence
              </button>
            )}
          </div>

          <div className="source-setting">
            <div className="source-header">
              <span className="source-name">Slack</span>
              <span className={`status ${getConnectionStatus(apiKeys.slack)}`}>
                {getConnectionStatus(apiKeys.slack)}
              </span>
            </div>
            {apiKeys.slack ? (
              <div className="connected-info">
                <span className="connected-text">✓ Connected to Slack</span>
                <button 
                  onClick={() => handleKeyChange('slack', '')}
                  className="disconnect-button"
                >
                  Disconnect
                </button>
              </div>
            ) : (
              <button 
                onClick={() => handleOAuthConnect('slack')}
                className="connect-button"
              >
                Connect to Slack
              </button>
            )}
          </div>

          <div className="source-setting">
            <div className="source-header">
              <span className="source-name">Gmail</span>
              <span className={`status ${getConnectionStatus(apiKeys.gmail)}`}>
                {getConnectionStatus(apiKeys.gmail)}
              </span>
            </div>
            {apiKeys.gmail ? (
              <div className="connected-info">
                <span className="connected-text">✓ Connected to Gmail</span>
                <button 
                  onClick={() => handleKeyChange('gmail', '')}
                  className="disconnect-button"
                >
                  Disconnect
                </button>
              </div>
            ) : (
              <button 
                onClick={() => handleOAuthConnect('gmail')}
                className="connect-button"
              >
                Connect to Gmail
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Settings;