package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ConfluenceService struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type ConfluenceOAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type ConfluenceResource struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ConfluenceResourcesResponse struct {
	Values []ConfluenceResource `json:"values"`
}

type ConfluenceSearchResult struct {
	Results []ConfluenceContent `json:"results"`
	Size    int                 `json:"size"`
}

type ConfluenceContent struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Links struct {
		Base   string `json:"base"`
		WebUI  string `json:"webui"`
		TinyUI string `json:"tinyui"`
	} `json:"_links"`
	Space struct {
		Name string `json:"name"`
	} `json:"space"`
	Body struct {
		View struct {
			Value string `json:"value"`
		} `json:"view"`
	} `json:"body"`
}

type ConfluenceContentDetail struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  struct {
		View struct {
			Value string `json:"value"`
		} `json:"view"`
		Storage struct {
			Value string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
	Space struct {
		Name string `json:"name"`
	} `json:"space"`
}

func NewConfluenceService(clientID, clientSecret, redirectURL string) *ConfluenceService {
	return &ConfluenceService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (cs *ConfluenceService) GetAuthURL(state string) string {
	baseURL := "https://auth.atlassian.com/authorize"
	params := url.Values{}
	params.Add("audience", "api.atlassian.com")
	params.Add("client_id", cs.ClientID)
	params.Add("scope", "read:confluence-content.all read:confluence-content.summary read:confluence-space.summary search:confluence")
	params.Add("redirect_uri", cs.RedirectURL)
	params.Add("state", state)
	params.Add("response_type", "code")
	params.Add("prompt", "consent")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func (cs *ConfluenceService) ExchangeCodeForToken(code string) (*ConfluenceOAuthResponse, error) {
	tokenURL := "https://auth.atlassian.com/oauth/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", cs.ClientID)
	data.Set("client_secret", cs.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", cs.RedirectURL)

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

	var tokenResponse ConfluenceOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func (cs *ConfluenceService) GetAccessibleResources(accessToken string) (*ConfluenceResourcesResponse, error) {
	resourcesURL := "https://api.atlassian.com/oauth/token/accessible-resources"

	req, err := http.NewRequest("GET", resourcesURL, nil)
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
		return nil, fmt.Errorf("failed to get accessible resources: %s", resp.Status)
	}

	var resources ConfluenceResourcesResponse
	if err := json.NewDecoder(resp.Body).Decode(&resources.Values); err != nil {
		return nil, err
	}

	return &resources, nil
}

func (cs *ConfluenceService) SearchContent(accessToken, query, cloudID string) (*ConfluenceSearchResult, error) {
	searchURL := fmt.Sprintf("https://api.atlassian.com/ex/confluence/%s/rest/api/content/search", cloudID)

	params := url.Values{}
	params.Add("cql", fmt.Sprintf("text ~ \"%s\"", query))
	params.Add("limit", "10")
	params.Add("expand", "space,body.view,body.storage")

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
		return nil, fmt.Errorf("failed to search Confluence: %s", resp.Status)
	}

	var searchResult ConfluenceSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	return &searchResult, nil
}

func (cs *ConfluenceService) GetContentDetail(accessToken, contentID, cloudID string) (*ConfluenceContentDetail, error) {
	contentURL := fmt.Sprintf("https://api.atlassian.com/ex/confluence/%s/rest/api/content/%s", cloudID, contentID)
	
	params := url.Values{}
	params.Add("expand", "body.view,body.storage,space")
	
	fullURL := fmt.Sprintf("%s?%s", contentURL, params.Encode())

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
		return nil, fmt.Errorf("failed to get content detail: %s", resp.Status)
	}

	var contentDetail ConfluenceContentDetail
	if err := json.NewDecoder(resp.Body).Decode(&contentDetail); err != nil {
		return nil, err
	}

	return &contentDetail, nil
}
