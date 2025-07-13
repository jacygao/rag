import React from 'react';
import { ChatMessage } from '../types';
import References from './References';
import './MessageBubble.css';

interface MessageBubbleProps {
  message: ChatMessage;
}

const MessageBubble: React.FC<MessageBubbleProps> = ({ message }) => {
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className={`message-bubble ${message.isUser ? 'user' : 'bot'}`} data-message-id={message.id}>
      <div className="message-content">
        <div className="message-text">{message.content}</div>
        {message.references && message.references.length > 0 && (
          <References references={message.references} />
        )}
      </div>
      <div className="message-time">{formatTime(message.timestamp)}</div>
    </div>
  );
};

export default MessageBubble;