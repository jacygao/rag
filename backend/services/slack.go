package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SlackService struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type SlackOAuthResponse struct {
	OK          bool   `json:"ok"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Enterprise interface{} `json:"enterprise"`
	AuthedUser struct {
		ID          string `json:"id"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	} `json:"authed_user"`
}

type SlackSearchResponse struct {
	OK       bool `json:"ok"`
	Query    string `json:"query"`
	Messages struct {
		Total      int            `json:"total"`
		Pagination SlackPagination `json:"pagination"`
		Paging     SlackPaging     `json:"paging"`
		Matches    []SlackMessage  `json:"matches"`
	} `json:"messages"`
}

type SlackPagination struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	PageCount  int `json:"page_count"`
	First      int `json:"first"`
	Last       int `json:"last"`
}

type SlackPaging struct {
	Count int `json:"count"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
}

type SlackMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	User      string `json:"user"`
	Username  string `json:"username"`
	Ts        string `json:"ts"`
	Team      string `json:"team"`
	Channel   SlackChannel `json:"channel"`
	Permalink string `json:"permalink"`
}

type SlackChannel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsChannel  bool   `json:"is_channel"`
	IsGroup    bool   `json:"is_group"`
	IsIM       bool   `json:"is_im"`
	IsMpim     bool   `json:"is_mpim"`
	IsPrivate  bool   `json:"is_private"`
	IsOrgShared bool  `json:"is_org_shared"`
}

func NewSlackService(clientID, clientSecret, redirectURL string) *SlackService {
	return &SlackService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (ss *SlackService) GetAuthURL(state string) string {
	baseURL := "https://slack.com/oauth/v2/authorize"
	params := url.Values{}
	params.Add("client_id", ss.ClientID)
	params.Add("user_scope", "search:read")
	params.Add("redirect_uri", ss.RedirectURL)
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func (ss *SlackService) ExchangeCodeForToken(code string) (*SlackOAuthResponse, error) {
	tokenURL := "https://slack.com/api/oauth.v2.access"

	data := url.Values{}
	data.Set("client_id", ss.ClientID)
	data.Set("client_secret", ss.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", ss.RedirectURL)

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

	var tokenResponse SlackOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}


	if !tokenResponse.OK {
		return nil, fmt.Errorf("slack oauth failed")
	}

	return &tokenResponse, nil
}

func (ss *SlackService) SearchMessages(accessToken, query string, count int) (*SlackSearchResponse, error) {
	if count == 0 {
		count = 10
	}

	searchURL := "https://slack.com/api/search.messages"
	params := url.Values{}
	params.Add("query", query)
	params.Add("count", fmt.Sprintf("%d", count))
	params.Add("sort", "timestamp")
	params.Add("sort_dir", "desc")

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())
	
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search Slack: %s", resp.Status)
	}

	var searchResults SlackSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		return nil, err
	}

	if !searchResults.OK {
		return nil, fmt.Errorf("slack search failed")
	}

	return &searchResults, nil
}

// Helper function to format timestamp
func (ss *SlackService) FormatSlackTimestamp(ts string) time.Time {
	// Slack timestamps are in format "1234567890.123456"
	// We'll parse just the seconds part for simplicity
	if len(ts) > 10 {
		ts = ts[:10]
	}
	
	// Parse as Unix timestamp
	timestamp := time.Unix(0, 0)
	if len(ts) >= 10 {
		// Simple parsing - in production you'd want better error handling
		timestamp = time.Unix(1234567890, 0) // Placeholder
	}
	
	return timestamp
}

// Helper function to clean Slack message text
func (ss *SlackService) CleanSlackText(text string) string {
	// Remove Slack formatting
	text = strings.ReplaceAll(text, "<@", "@")
	text = strings.ReplaceAll(text, ">", "")
	text = strings.ReplaceAll(text, "<#", "#")
	text = strings.ReplaceAll(text, "<!", "")
	
	// Remove URLs in format <http://example.com|text>
	text = strings.ReplaceAll(text, "<http", "http")
	text = strings.ReplaceAll(text, "<https", "https")
	
	return text
}