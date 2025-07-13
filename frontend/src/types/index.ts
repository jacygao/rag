export interface ApiKeys {
  confluence: string;
  slack: string;
  gmail: string;
}

export interface ChatMessage {
  id: string;
  content: string;
  isUser: boolean;
  timestamp: Date;
  references?: Reference[];
  isStreaming?: boolean;
}

export interface Reference {
  title: string;
  url: string;
  source: string;
}

export interface ChatRequest {
  query: string;
  confluence_token?: string;
  slack_token?: string;
  gmail_token?: string;
  sources: Record<string, string>;
}

export interface ChatResponse {
  response: string;
  references: Reference[];
}