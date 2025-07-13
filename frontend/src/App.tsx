import React, { useState } from 'react';
import Chat from './components/Chat';
import Settings from './components/Settings';
import { ApiKeys } from './types';
import './App.css';


const App: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<ApiKeys>({
    confluence: sessionStorage.getItem('confluence_token') || '',
    slack: sessionStorage.getItem('slack_token') || '',
    gmail: sessionStorage.getItem('gmail_token') || ''
  });

  const updateApiKeys = (keys: Partial<ApiKeys>) => {
    const newKeys = { ...apiKeys, ...keys };
    setApiKeys(newKeys);
    
    Object.entries(keys).forEach(([key, value]) => {
      if (value) {
        sessionStorage.setItem(`${key}_token`, value);
      } else {
        sessionStorage.removeItem(`${key}_token`);
      }
    });
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>RAG Chatbot</h1>
      </header>
      
      <main className="app-main">
        <Settings apiKeys={apiKeys} onUpdateKeys={updateApiKeys} />
        <Chat apiKeys={apiKeys} />
      </main>
    </div>
  );
};

export default App;