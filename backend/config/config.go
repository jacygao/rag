package config

import (
	"log"
	"os"
)

type Config struct {
	Port        string
	FrontendURL string
	
	Confluence struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	
	Slack struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	
	Google struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	
	OpenAI struct {
		APIKey string
		Model  string
	}
}

func Load() *Config {
	config := &Config{
		Port:        getEnv("APP_PORT", "8085"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	// Determine protocol based on HTTPS setting
	protocol := "http"
	if getEnv("USE_HTTPS", "false") == "true" {
		protocol = "https"
	}

	// Confluence OAuth
	config.Confluence.ClientID = getEnv("CONFLUENCE_CLIENT_ID", "")
	config.Confluence.ClientSecret = getEnv("CONFLUENCE_CLIENT_SECRET", "")
	config.Confluence.RedirectURL = protocol + "://localhost:" + config.Port + "/api/auth/confluence/callback"

	// Slack OAuth
	config.Slack.ClientID = getEnv("SLACK_CLIENT_ID", "")
	config.Slack.ClientSecret = getEnv("SLACK_CLIENT_SECRET", "")
	config.Slack.RedirectURL = protocol + "://localhost:" + config.Port + "/api/auth/slack/callback"

	// Google OAuth
	config.Google.ClientID = getEnv("GOOGLE_CLIENT_ID", "")
	config.Google.ClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")
	config.Google.RedirectURL = protocol + "://localhost:" + config.Port + "/api/auth/google/callback"

	// OpenAI
	config.OpenAI.APIKey = getEnv("OPENAI_API_KEY", "")
	config.OpenAI.Model = getEnv("OPENAI_MODEL", "gpt-4")

	if config.Confluence.ClientID == "" {
		log.Println("Warning: CONFLUENCE_CLIENT_ID not set")
	}
	
	if config.OpenAI.APIKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}