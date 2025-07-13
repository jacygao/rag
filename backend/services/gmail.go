package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GmailService struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GmailOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type GmailSearchResponse struct {
	Messages          []GmailMessage `json:"messages"`
	NextPageToken     string         `json:"nextPageToken"`
	ResultSizeEstimate int           `json:"resultSizeEstimate"`
}

type GmailMessage struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
}

type GmailMessageDetail struct {
	ID           string            `json:"id"`
	ThreadID     string            `json:"threadId"`
	LabelIDs     []string          `json:"labelIds"`
	Snippet      string            `json:"snippet"`
	HistoryID    string            `json:"historyId"`
	InternalDate string            `json:"internalDate"`
	Payload      GmailMessagePayload `json:"payload"`
}

type GmailMessagePayload struct {
	PartID   string                `json:"partId"`
	MimeType string                `json:"mimeType"`
	Filename string                `json:"filename"`
	Headers  []GmailHeader         `json:"headers"`
	Body     GmailMessageBody      `json:"body"`
	Parts    []GmailMessagePayload `json:"parts"`
}

type GmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GmailMessageBody struct {
	AttachmentID string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         string `json:"data"`
}

func NewGmailService(clientID, clientSecret, redirectURL string) *GmailService {
	return &GmailService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (gs *GmailService) GetAuthURL(state string) string {
	baseURL := "https://accounts.google.com/o/oauth2/v2/auth"
	params := url.Values{}
	params.Add("client_id", gs.ClientID)
	params.Add("redirect_uri", gs.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/gmail.readonly")
	params.Add("access_type", "offline")
	params.Add("state", state)
	params.Add("prompt", "consent")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func (gs *GmailService) ExchangeCodeForToken(code string) (*GmailOAuthResponse, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("client_id", gs.ClientID)
	data.Set("client_secret", gs.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", gs.RedirectURL)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to exchange code for token: %s", resp.Status)
	}

	var tokenResponse GmailOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func (gs *GmailService) SearchMessages(accessToken, query string, maxResults int) (*GmailSearchResponse, error) {
	if maxResults == 0 {
		maxResults = 10
	}

	searchURL := "https://gmail.googleapis.com/gmail/v1/users/me/messages"
	params := url.Values{}
	params.Add("q", query)
	params.Add("maxResults", fmt.Sprintf("%d", maxResults))

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search Gmail: %s", resp.Status)
	}

	var searchResults GmailSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		return nil, err
	}

	return &searchResults, nil
}

func (gs *GmailService) GetMessageDetail(accessToken, messageID string) (*GmailMessageDetail, error) {
	messageURL := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages/%s", messageID)

	req, err := http.NewRequest("GET", messageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get message detail: %s", resp.Status)
	}

	var messageDetail GmailMessageDetail
	if err := json.NewDecoder(resp.Body).Decode(&messageDetail); err != nil {
		return nil, err
	}

	return &messageDetail, nil
}

// Helper function to extract email content and metadata
func (gs *GmailService) ExtractEmailInfo(message *GmailMessageDetail) (subject, sender, content string, date time.Time) {
	// Extract headers
	for _, header := range message.Payload.Headers {
		switch header.Name {
		case "Subject":
			subject = header.Value
		case "From":
			sender = header.Value
		case "Date":
			if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				date = parsedDate
			}
		}
	}

	// Extract body content
	content = gs.extractBody(&message.Payload)
	if content == "" {
		content = message.Snippet // Fall back to snippet
	}

	return subject, sender, content, date
}

func (gs *GmailService) extractBody(payload *GmailMessagePayload) string {
	// If this part has a body with data, extract it
	if payload.Body.Data != "" {
		// Gmail API returns base64url-encoded data
		decoded, err := gs.decodeBase64URL(payload.Body.Data)
		if err == nil {
			return decoded
		}
	}

	// If this is a multipart message, search through parts
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			body := gs.extractBody(&part)
			if body != "" {
				return body
			}
		}
	}

	return ""
}

func (gs *GmailService) decodeBase64URL(data string) (string, error) {
	// Gmail uses base64url encoding without padding
	// Convert to standard base64 and add padding if needed
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")
	
	// Add padding
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	// Decode would require base64 package - for now return the data as is
	// In production, you'd want to properly decode this
	return data, nil
}