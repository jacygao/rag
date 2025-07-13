package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIService struct {
	APIKey string
	Model  string
}

type OpenAIRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIMessage     `json:"messages"`
	Temperature float64             `json:"temperature"`
	MaxTokens   int                 `json:"max_tokens"`
	Stream      bool                `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Streaming response types
type OpenAIStreamResponse struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []OpenAIStreamChoice  `json:"choices"`
}

type OpenAIStreamChoice struct {
	Index int                 `json:"index"`
	Delta OpenAIStreamDelta   `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

type OpenAIStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

func NewOpenAIService(apiKey, model string) *OpenAIService {
	if model == "" {
		model = "gpt-4" // Default to GPT-4 for enterprise account
	}
	return &OpenAIService{
		APIKey: apiKey,
		Model:  model,
	}
}

func (ai *OpenAIService) GenerateResponse(userQuery string, searchResults []SearchResult) (*OpenAIResponse, error) {
	// Build context from search results
	var contextParts []string
	for _, result := range searchResults {
		contextParts = append(contextParts, fmt.Sprintf("From %s (%s): %s", result.Title, result.Source, result.Content))
	}
	context := strings.Join(contextParts, "\n\n")

	// Create system prompt for RAG
	systemPrompt := `You are a helpful AI assistant that answers questions based on the provided context from the user's work documents. 

Instructions:
1. Answer the user's question using ONLY the information provided in the context
2. If the context doesn't contain relevant information, say so clearly
3. Be concise but thorough in your response
4. Reference which sources you're drawing from when relevant
5. If you're unsure about something, acknowledge the uncertainty

Context from user's documents:
` + context

	// Create the request
	request := OpenAIRequest{
		Model: ai.Model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userQuery,
			},
		},
		Temperature: 0.3, // Lower temperature for more focused responses
		MaxTokens:   1000,
		Stream:      false,
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.APIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", resp.Status)
	}

	// Parse response
	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &openaiResp, nil
}

func (ai *OpenAIService) GenerateStreamingResponse(userQuery string, searchResults []SearchResult, writer io.Writer) error {
	// Build context from search results
	var contextParts []string
	for _, result := range searchResults {
		contextParts = append(contextParts, fmt.Sprintf("From %s (%s): %s", result.Title, result.Source, result.Content))
	}
	context := strings.Join(contextParts, "\n\n")

	// Create system prompt for RAG
	systemPrompt := `You are a helpful AI assistant that answers questions based on the provided context from the user's work documents. 

Instructions:
1. Answer the user's question using ONLY the information provided in the context
2. If the context doesn't contain relevant information, say so clearly
3. Be concise but thorough in your response
4. Reference which sources you're drawing from when relevant
5. If you're unsure about something, acknowledge the uncertainty

Context from user's documents:
` + context

	// Create the request with streaming enabled
	request := OpenAIRequest{
		Model: ai.Model,
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userQuery,
			},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
		Stream:      true, // Enable streaming
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.APIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API error: %s", resp.Status)
	}

	// Process streaming response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Debug logging
		fmt.Printf("OpenAI stream line: %s\n", line)
		
		// Skip empty lines
		if len(line) == 0 {
			continue
		}
		
		// OpenAI streaming format: "data: {json}" or "data: [DONE]"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		
		// Remove "data: " prefix
		data := strings.TrimPrefix(line, "data: ")
		
		// Check for completion
		if data == "[DONE]" {
			fmt.Printf("OpenAI stream completed\n")
			// Send completion event
			fmt.Fprintf(writer, "data: {\"type\":\"done\"}\n\n")
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
			break
		}
		
		// Parse the streaming response
		var streamResp OpenAIStreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			fmt.Printf("Failed to parse OpenAI stream data: %v\n", err)
			continue // Skip malformed responses
		}
		
		// Extract content from the response
		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]
			if choice.Delta.Content != "" {
				fmt.Printf("Streaming content chunk: %s\n", choice.Delta.Content)
				// Send content chunk
				chunkData := map[string]string{
					"type":    "content",
					"content": choice.Delta.Content,
				}
				chunkJSON, _ := json.Marshal(chunkData)
				fmt.Fprintf(writer, "data: %s\n\n", chunkJSON)
				
				// Flush immediately for real-time streaming
				if flusher, ok := writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}
	
	return nil
}

// SearchResult represents a search result from any source
type SearchResult struct {
	Title   string
	Content string
	Source  string
	URL     string
}