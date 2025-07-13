package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rag-chatbot/config"
	"rag-chatbot/services"
)

var (
	confluenceService *services.ConfluenceService
	gmailService      *services.GmailService
	slackService      *services.SlackService
	appConfig         *config.Config
)

func init() {
	appConfig = config.Load()
	confluenceService = services.NewConfluenceService(
		appConfig.Confluence.ClientID,
		appConfig.Confluence.ClientSecret,
		appConfig.Confluence.RedirectURL,
	)
	gmailService = services.NewGmailService(
		appConfig.Google.ClientID,
		appConfig.Google.ClientSecret,
		appConfig.Google.RedirectURL,
	)
	slackService = services.NewSlackService(
		appConfig.Slack.ClientID,
		appConfig.Slack.ClientSecret,
		appConfig.Slack.RedirectURL,
	)
}

type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func ConfluenceAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := "random-state-" + "12345" // TODO: Generate secure random state
	authURL := confluenceService.GetAuthURL(state)

	response := AuthURLResponse{
		AuthURL: authURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GmailAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if appConfig.Google.ClientID == "" {
		http.Error(w, "Gmail OAuth not configured: GOOGLE_CLIENT_ID not set", http.StatusInternalServerError)
		return
	}

	state := "gmail-state-" + "12345" // TODO: Generate secure random state
	authURL := gmailService.GetAuthURL(state)

	response := AuthURLResponse{
		AuthURL: authURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GmailCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// TODO: Verify state parameter matches
	_ = state

	tokenResponse, err := gmailService.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTML page that posts the token back to the parent window
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Gmail Authorization Complete</title>
</head>
<body>
    <script>
        // Send token to parent window and close popup
        if (window.opener) {
            window.opener.postMessage({
                type: 'GMAIL_AUTH_SUCCESS',
                token: '` + tokenResponse.AccessToken + `',
                expiresIn: ` + fmt.Sprintf("%d", tokenResponse.ExpiresIn) + `
            }, '*');
            
            setTimeout(() => {
                window.close();
            }, 500);
        } else {
            document.body.innerHTML = '<h2>Authorization successful! You can close this window.</h2>';
        }
    </script>
    <h2>Connecting to Gmail...</h2>
    <p>This window should close automatically.</p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func SlackAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if appConfig.Slack.ClientID == "" {
		http.Error(w, "Slack OAuth not configured: SLACK_CLIENT_ID not set", http.StatusInternalServerError)
		return
	}

	state := "slack-state-" + "12345" // TODO: Generate secure random state
	authURL := slackService.GetAuthURL(state)

	response := AuthURLResponse{
		AuthURL: authURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SlackCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// TODO: Verify state parameter matches
	_ = state

	tokenResponse, err := slackService.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	
	if !tokenResponse.OK {
		http.Error(w, "Slack OAuth failed", http.StatusInternalServerError)
		return
	}

	// For user scopes, the access token is in authed_user.access_token
	userAccessToken := tokenResponse.AuthedUser.AccessToken
	if userAccessToken == "" {
		http.Error(w, "Empty user access token received", http.StatusInternalServerError)
		return
	}

	// Return HTML page that posts the token back to the parent window
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Slack Authorization Complete</title>
</head>
<body>
    <script>
        // Send token to parent window and close popup
        if (window.opener) {
            window.opener.postMessage({
                type: 'SLACK_AUTH_SUCCESS',
                token: '` + userAccessToken + `'
            }, '*');
            
            setTimeout(() => {
                window.close();
            }, 500);
        } else {
            document.body.innerHTML = '<h2>Authorization successful! You can close this window.</h2>';
        }
    </script>
    <h2>Connecting to Slack...</h2>
    <p>This window should close automatically.</p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func ConfluenceCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// TODO: Verify state parameter matches
	_ = state // Ignore for now, will implement state verification later

	tokenResponse, err := confluenceService.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return HTML page that posts the token back to the parent window
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Complete</title>
</head>
<body>
    <script>
        // Send token to parent window and close popup
        if (window.opener) {
            window.opener.postMessage({
                type: 'CONFLUENCE_AUTH_SUCCESS',
                token: '` + tokenResponse.AccessToken + `',
                expiresIn: ` + fmt.Sprintf("%d", tokenResponse.ExpiresIn) + `
            }, '*');
            
            setTimeout(() => {
                window.close();
            }, 500);
        } else {
            document.body.innerHTML = '<h2>Authorization successful! You can close this window.</h2>';
        }
    </script>
    <h2>Connecting to Confluence...</h2>
    <p>This window should close automatically.</p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}