import React, { useState, useRef, useEffect } from 'react';
import { ApiKeys, ChatMessage, Reference } from '../types';
import MessageBubble from './MessageBubble';
import './Chat.css';

interface ChatProps {
  apiKeys: ApiKeys;
}

const Chat: React.FC<ChatProps> = ({ apiKeys }) => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const streamingMessageRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const sendMessage = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      content: input.trim(),
      isUser: true,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    // Create a placeholder message for streaming
    const botMessageId = (Date.now() + 1).toString();
    const botMessage: ChatMessage = {
      id: botMessageId,
      content: '',
      isUser: false,
      timestamp: new Date(),
      references: [],
      isStreaming: true, // Mark as streaming for special handling
    };
    setMessages(prev => [...prev, botMessage]);

    try {
      const response = await fetch('/api/chat/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: userMessage.content,
          confluence_token: apiKeys.confluence,
          slack_token: apiKeys.slack,
          gmail_token: apiKeys.gmail,
          sources: {
            confluence: apiKeys.confluence ? 'enabled' : 'disabled',
            slack: apiKeys.slack ? 'enabled' : 'disabled',
            gmail: apiKeys.gmail ? 'enabled' : 'disabled',
          },
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to get response');
      }

      const reader = response.body?.getReader();
      const decoder = new TextDecoder();

      if (!reader) {
        throw new Error('No response body');
      }

      let buffer = '';
      
      while (true) {
        const { done, value } = await reader.read();
        
        if (done) break;
        
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        
        // Keep the last line in buffer (might be incomplete)
        buffer = lines.pop() || '';
        
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            
            if (data.trim() === '') continue;
            
            try {
              const event = JSON.parse(data);
              console.log('Received SSE event:', event);
              
              if (event.type === 'content') {
                console.log('Adding content chunk:', event.content);
                
                // Update DOM directly for immediate visual feedback
                const streamingElement = document.querySelector(`[data-message-id="${botMessageId}"] .message-text`);
                if (streamingElement) {
                  streamingElement.textContent += event.content;
                  // Scroll to bottom after each update
                  scrollToBottom();
                }
                
                // Also update React state (but don't rely on it for immediate rendering)
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === botMessageId
                      ? { ...msg, content: msg.content + event.content }
                      : msg
                  )
                );
              } else if (event.type === 'references') {
                console.log('Setting references:', event.references);
                // Set references
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === botMessageId
                      ? { ...msg, references: event.references }
                      : msg
                  )
                );
              } else if (event.type === 'done') {
                console.log('Stream completed');
                setIsLoading(false);
                // Mark the message as no longer streaming
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === botMessageId
                      ? { ...msg, isStreaming: false }
                      : msg
                  )
                );
                break;
              } else if (event.type === 'error') {
                console.log('Stream error:', event.message);
                setMessages(prev => 
                  prev.map(msg => 
                    msg.id === botMessageId
                      ? { ...msg, content: event.message }
                      : msg
                  )
                );
                setIsLoading(false);
                break;
              } else if (event.type === 'status') {
                console.log('Status update:', event.message);
              }
            } catch (e) {
              console.log('Failed to parse SSE data:', data, e);
              // Skip malformed JSON
              continue;
            }
          }
        }
      }
    } catch (error) {
      setMessages(prev => 
        prev.map(msg => 
          msg.id === botMessageId
            ? { ...msg, content: 'Sorry, I encountered an error. Please try again.' }
            : msg
        )
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const hasActiveConnections = Object.values(apiKeys).some(key => key.length > 0);

  return (
    <div className="chat">
      <div className="chat-header">
        <h2>Chat</h2>
        <div className="active-sources">
          {apiKeys.confluence && <span className="source-badge">Confluence</span>}
          {apiKeys.slack && <span className="source-badge">Slack</span>}
          {apiKeys.gmail && <span className="source-badge">Gmail</span>}
          {!hasActiveConnections && <span className="no-sources">No data sources connected</span>}
        </div>
      </div>

      <div className="messages-container">
        {messages.length === 0 && (
          <div className="welcome-message">
            <h3>Welcome to RAG Chatbot!</h3>
            <p>Ask me anything about your work. I'll search through your connected data sources to find relevant information.</p>
            {!hasActiveConnections && (
              <p className="setup-prompt">
                👈 Connect your data sources in the sidebar to get started.
              </p>
            )}
          </div>
        )}
        
        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
        
        {isLoading && (
          <div className="typing-indicator">
            <div className="typing-dots">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <span className="typing-text">Searching and thinking...</span>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      <div className="input-container">
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder={hasActiveConnections ? "Ask me anything..." : "Connect data sources to start chatting"}
          className="message-input"
          rows={1}
          disabled={!hasActiveConnections || isLoading}
        />
        <button
          onClick={sendMessage}
          disabled={!input.trim() || !hasActiveConnections || isLoading}
          className="send-button"
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default Chat;