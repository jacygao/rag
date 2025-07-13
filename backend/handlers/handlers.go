package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"rag-chatbot/config"
	"rag-chatbot/services"
)

var (
	openaiService  *services.OpenAIService
	rankingService *services.RankingService
)

func init() {
	cfg := config.Load()
	openaiService = services.NewOpenAIService(
		cfg.OpenAI.APIKey,
		cfg.OpenAI.Model,
	)
	rankingService = services.NewRankingService()
}


type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

type ChatRequest struct {
	Query           string            `json:"query"`
	ConfluenceToken string            `json:"confluence_token,omitempty"`
	SlackToken      string            `json:"slack_token,omitempty"`
	GmailToken      string            `json:"gmail_token,omitempty"`
	Sources         map[string]string `json:"sources"`
}

type ChatResponse struct {
	Response   string      `json:"response"`
	References []Reference `json:"references"`
}

type Reference struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Source string `json:"source"`
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var allReferences []Reference
	var allSearchResults []services.SearchResult

	// Search Confluence if token is provided
	if req.ConfluenceToken != "" {
		confluenceResults, err := searchConfluence(req.ConfluenceToken, req.Query)
		if err != nil {
			log.Printf("Confluence search error: %v", err)
		} else {
			allReferences = append(allReferences, confluenceResults.References...)
			// Use the enhanced search results with actual content
			allSearchResults = append(allSearchResults, confluenceResults.SearchResults...)
		}
	}

	// Search Gmail if token is provided
	if req.GmailToken != "" {
		log.Printf("Gmail token provided, searching for: %s", req.Query)
		gmailResults, err := searchGmail(req.GmailToken, req.Query)
		if err != nil {
			log.Printf("Gmail search error: %v", err)
		} else {
			log.Printf("Gmail search returned %d references and %d search results", len(gmailResults.References), len(gmailResults.SearchResults))
			allReferences = append(allReferences, gmailResults.References...)
			allSearchResults = append(allSearchResults, gmailResults.SearchResults...)
		}
	}

	// Search Slack if token is provided
	if req.SlackToken != "" {
		log.Printf("Slack token provided, searching for: %s", req.Query)
		slackResults, err := searchSlack(req.SlackToken, req.Query)
		if err != nil {
			log.Printf("Slack search error: %v", err)
		} else {
			log.Printf("Slack search returned %d references and %d search results", len(slackResults.References), len(slackResults.SearchResults))
			allReferences = append(allReferences, slackResults.References...)
			allSearchResults = append(allSearchResults, slackResults.SearchResults...)
		}
	}

	// Generate response using OpenAI
	var responseText string
	if len(allSearchResults) > 0 {
		aiResponse, err := openaiService.GenerateResponse(req.Query, allSearchResults)
		if err != nil {
			log.Printf("OpenAI error: %v", err)
			responseText = fmt.Sprintf("Found %d relevant results from your sources, but couldn't generate a detailed response. Please try again.", len(allSearchResults))
		} else if len(aiResponse.Choices) > 0 {
			responseText = aiResponse.Choices[0].Message.Content
		} else {
			responseText = "No response generated from AI service."
		}
	} else {
		responseText = "I couldn't find any relevant information in your connected sources. Please make sure you've connected your data sources and try a different query."
	}

	response := ChatResponse{
		Response:   responseText,
		References: allReferences,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ChatStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Send initial status
	fmt.Printf("Starting SSE stream\n")
	fmt.Fprintf(w, "data: {\"type\":\"status\",\"message\":\"Searching your sources...\"}\n\n")
	w.(http.Flusher).Flush()

	var allReferences []Reference
	var allSearchResults []services.SearchResult

	// Search Confluence if token is provided
	if req.ConfluenceToken != "" {
		confluenceResults, err := searchConfluence(req.ConfluenceToken, req.Query)
		if err != nil {
			log.Printf("Confluence search error: %v", err)
		} else {
			allReferences = append(allReferences, confluenceResults.References...)
			allSearchResults = append(allSearchResults, confluenceResults.SearchResults...)
		}
	}

	// Search Gmail if token is provided
	if req.GmailToken != "" {
		gmailResults, err := searchGmail(req.GmailToken, req.Query)
		if err != nil {
			log.Printf("Gmail search error: %v", err)
		} else {
			allReferences = append(allReferences, gmailResults.References...)
			allSearchResults = append(allSearchResults, gmailResults.SearchResults...)
		}
	}

	// Search Slack if token is provided
	if req.SlackToken != "" {
		slackResults, err := searchSlack(req.SlackToken, req.Query)
		if err != nil {
			log.Printf("Slack search error: %v", err)
		} else {
			allReferences = append(allReferences, slackResults.References...)
			allSearchResults = append(allSearchResults, slackResults.SearchResults...)
		}
	}

	// Send references
	if len(allReferences) > 0 {
		referencesData := map[string]interface{}{
			"type":       "references",
			"references": allReferences,
		}
		referencesJSON, _ := json.Marshal(referencesData)
		fmt.Fprintf(w, "data: %s\n\n", referencesJSON)
		w.(http.Flusher).Flush()
	}

	// Send status update
	fmt.Fprintf(w, "data: {\"type\":\"status\",\"message\":\"Generating response...\"}\n\n")
	w.(http.Flusher).Flush()

	// Generate streaming response
	if len(allSearchResults) > 0 {
		err := openaiService.GenerateStreamingResponse(req.Query, allSearchResults, w)
		if err != nil {
			log.Printf("OpenAI streaming error: %v", err)
			errorData := map[string]string{
				"type":    "error",
				"message": "Failed to generate response. Please try again.",
			}
			errorJSON, _ := json.Marshal(errorData)
			fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		}
	} else {
		// Send no results message
		messageData := map[string]string{
			"type":    "content",
			"content": "I couldn't find any relevant information in your connected sources. Please make sure you've connected your data sources and try a different query.",
		}
		messageJSON, _ := json.Marshal(messageData)
		fmt.Fprintf(w, "data: %s\n\n", messageJSON)
		fmt.Fprintf(w, "data: {\"type\":\"done\"}\n\n")
	}

	w.(http.Flusher).Flush()
}

type ConfluenceSearchResults struct {
	References    []Reference
	SearchResults []services.SearchResult
}

func searchConfluence(accessToken, query string) (*ConfluenceSearchResults, error) {
	// Get accessible resources first (using the global confluenceService from oauth.go)
	resources, err := confluenceService.GetAccessibleResources(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible resources: %v", err)
	}

	if len(resources.Values) == 0 {
		return nil, fmt.Errorf("no accessible Confluence sites found")
	}

	// Use the first available site
	cloudID := resources.Values[0].ID
	baseURL := strings.TrimSuffix(resources.Values[0].URL, "/")

	// Search for content
	searchResults, err := confluenceService.SearchContent(accessToken, query, cloudID)
	if err != nil {
		return nil, fmt.Errorf("failed to search Confluence: %v", err)
	}

	if len(searchResults.Results) == 0 {
		return &ConfluenceSearchResults{
			References: []Reference{},
			SearchResults: []services.SearchResult{},
		}, nil
	}

	// Step 1: Convert search results to SearchResult format with available content
	var initialResults []services.SearchResult
	for _, content := range searchResults.Results {
		fullURL := baseURL + content.Links.WebUI
		
		// Extract content from the search result if available
		var contentText string
		if content.Body.View.Value != "" {
			// Extract plain text from HTML
			plainText := services.ExtractPlainText(content.Body.View.Value)
			// Use intelligent chunking to extract relevant sections
			contentText = services.ExtractRelevantSections(plainText, query, 1500)
		} else {
			contentText = content.Title // Fall back to title if no content
		}
		
		initialResults = append(initialResults, services.SearchResult{
			Title:   content.Title,
			Content: contentText,
			Source:  "confluence",
			URL:     fullURL,
		})
	}

	// Step 2: Rerank results to get top 3 most relevant (now with actual content)
	topResults := rankingService.RerankResults(query, initialResults, 3)

	// Step 3: Prepare final results and references
	var enhancedResults []services.SearchResult
	var references []Reference

	for i, result := range topResults {
		enhancedResults = append(enhancedResults, result)
		
		references = append(references, Reference{
			Title:  result.Title,
			URL:    result.URL,
			Source: "confluence",
		})

		// Limit to 3 to control token usage
		if i >= 2 {
			break
		}
	}

	return &ConfluenceSearchResults{
		References:    references,
		SearchResults: enhancedResults,
	}, nil
}

type GmailSearchResults struct {
	References    []Reference
	SearchResults []services.SearchResult
}

func searchGmail(accessToken, query string) (*GmailSearchResults, error) {
	log.Printf("Searching Gmail with query: %s", query)
	// Create Gmail service instance (reuse from oauth.go)
	searchResults, err := gmailService.SearchMessages(accessToken, query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search Gmail: %v", err)
	}

	log.Printf("Gmail API returned %d messages", len(searchResults.Messages))
	if len(searchResults.Messages) == 0 {
		log.Printf("No Gmail messages found for query: %s", query)
		return &GmailSearchResults{
			References:    []Reference{},
			SearchResults: []services.SearchResult{},
		}, nil
	}

	// Step 1: Get details for all messages to prepare for reranking
	var initialResults []services.SearchResult
	for _, message := range searchResults.Messages {
		messageDetail, err := gmailService.GetMessageDetail(accessToken, message.ID)
		if err != nil {
			log.Printf("Failed to get Gmail message detail for %s: %v", message.ID, err)
			continue
		}

		// Extract email information
		subject, sender, content, date := gmailService.ExtractEmailInfo(messageDetail)
		
		// Create a Gmail URL (this would need to be adjusted for actual Gmail links)
		gmailURL := fmt.Sprintf("https://mail.google.com/mail/u/0/#inbox/%s", message.ID)
		
		// Use intelligent chunking to extract relevant content
		relevantContent := services.ExtractRelevantSections(content, query, 1200) // Slightly smaller for emails
		
		// Combine subject and content for better context
		fullContent := fmt.Sprintf("Subject: %s\nFrom: %s\nDate: %s\n\n%s", 
			subject, sender, date.Format("2006-01-02 15:04"), relevantContent)

		initialResults = append(initialResults, services.SearchResult{
			Title:   subject,
			Content: fullContent,
			Source:  "gmail",
			URL:     gmailURL,
		})
	}

	// Step 2: Rerank to get top 3 most relevant emails
	topResults := rankingService.RerankResults(query, initialResults, 3)

	// Step 3: Prepare final results and references
	var enhancedResults []services.SearchResult
	var references []Reference

	for i, result := range topResults {
		enhancedResults = append(enhancedResults, result)
		
		references = append(references, Reference{
			Title:  result.Title,
			URL:    result.URL,
			Source: "gmail",
		})

		// Limit to 3 to control token usage
		if i >= 2 {
			break
		}
	}

	return &GmailSearchResults{
		References:    references,
		SearchResults: enhancedResults,
	}, nil
}

type SlackSearchResults struct {
	References    []Reference
	SearchResults []services.SearchResult
}

func searchSlack(accessToken, query string) (*SlackSearchResults, error) {
	log.Printf("Searching Slack with query: %s", query)
	searchResults, err := slackService.SearchMessages(accessToken, query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search Slack: %v", err)
	}

	log.Printf("Slack API returned %d messages", len(searchResults.Messages.Matches))
	if len(searchResults.Messages.Matches) == 0 {
		log.Printf("No Slack messages found for query: %s", query)
		return &SlackSearchResults{
			References:    []Reference{},
			SearchResults: []services.SearchResult{},
		}, nil
	}

	// Step 1: Convert slack messages to SearchResult format
	var initialResults []services.SearchResult
	for _, message := range searchResults.Messages.Matches {
		// Clean Slack formatting from message text
		cleanText := slackService.CleanSlackText(message.Text)
		
		// Use intelligent chunking to extract relevant content
		relevantContent := services.ExtractRelevantSections(cleanText, query, 1000) // Smaller for Slack messages
		
		// Format message with context
		channelInfo := message.Channel.Name
		if channelInfo == "" {
			channelInfo = "Direct Message"
		}
		
		formattedContent := fmt.Sprintf("Channel: #%s\nUser: %s\n\n%s", 
			channelInfo, message.Username, relevantContent)

		// Use permalink as URL, or construct one if not available
		messageURL := message.Permalink
		if messageURL == "" {
			messageURL = fmt.Sprintf("https://slack.com/app_redirect?channel=%s", message.Channel.ID)
		}

		initialResults = append(initialResults, services.SearchResult{
			Title:   fmt.Sprintf("Message in #%s", channelInfo),
			Content: formattedContent,
			Source:  "slack",
			URL:     messageURL,
		})
	}

	// Step 2: Rerank to get top 3 most relevant messages
	topResults := rankingService.RerankResults(query, initialResults, 3)

	// Step 3: Prepare final results and references
	var enhancedResults []services.SearchResult
	var references []Reference

	for i, result := range topResults {
		enhancedResults = append(enhancedResults, result)
		
		references = append(references, Reference{
			Title:  result.Title,
			URL:    result.URL,
			Source: "slack",
		})

		// Limit to 3 to control token usage
		if i >= 2 {
			break
		}
	}

	return &SlackSearchResults{
		References:    references,
		SearchResults: enhancedResults,
	}, nil
}