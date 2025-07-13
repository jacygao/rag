# OAuth Setup Guide

This guide helps you set up OAuth applications for Confluence, Slack, and Gmail integration.

## Confluence OAuth Setup

1. **Go to Atlassian Developer Console**
   - Visit: https://developer.atlassian.com/console/myapps/
   - Sign in with your Atlassian account

2. **Create a New App**
   - Click "Create" → "OAuth 2.0 integration"
   - Enter app name: "RAG Chatbot"
   - Accept the terms

3. **Configure OAuth Settings**
   - **Authorization callback URL**: `http://localhost:8085/api/auth/confluence/callback`
   - **Scopes**: Add these permissions:
     - `read:confluence-content.summary`
     - `read:confluence-space.summary`

4. **Get Your Credentials**
   - Copy the **Client ID**
   - Copy the **Client secret**
   - Add them to your `.env` file

## Gmail OAuth Setup

1. **Go to Google Cloud Console**
   - Visit: https://console.cloud.google.com/
   - Create a new project or select existing one

2. **Enable Gmail API**
   - Go to "APIs & Services" → "Library"
   - Search for "Gmail API" and enable it

3. **Create OAuth 2.0 Credentials**
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth 2.0 Client IDs"
   - Application type: "Web application"
   - Name: "RAG Chatbot"
   - Authorized redirect URIs: `http://localhost:8085/api/auth/gmail/callback`

4. **Get Your Credentials**
   - Copy the **Client ID**
   - Copy the **Client secret**
   - Add them to your `.env` file

## Slack OAuth Setup

1. **Go to Slack API**
   - Visit: https://api.slack.com/apps
   - Click "Create New App" → "From scratch"
   - App Name: "RAG Chatbot"
   - Choose your workspace

2. **Configure OAuth & Permissions**
   - Go to "OAuth & Permissions" in the sidebar
   - Add redirect URL: `https://localhost:8085/api/auth/slack/callback`
   - Under "Scopes" → "User Token Scopes" (NOT Bot Token Scopes), add: `search:read`
   - Important: Make sure you add the scope under "User Token Scopes", not "Bot Token Scopes"

3. **Get Your Credentials**
   - Go to "Basic Information" in sidebar
   - Copy the **Client ID** 
   - Copy the **Client Secret**
   - Add them to your `.env` file

## Environment Setup

1. **Copy the example environment file:**
   ```bash
   cp backend/.env.example backend/.env
   ```

2. **Edit the `.env` file:**
   ```env
   CONFLUENCE_CLIENT_ID=your-actual-client-id-here
   CONFLUENCE_CLIENT_SECRET=your-actual-client-secret-here
   GOOGLE_CLIENT_ID=your-actual-google-client-id-here
   GOOGLE_CLIENT_SECRET=your-actual-google-client-secret-here
   SLACK_CLIENT_ID=your-actual-slack-client-id-here
   SLACK_CLIENT_SECRET=your-actual-slack-client-secret-here
   OPENAI_API_KEY=your-actual-openai-api-key-here
   ```

3. **Important Notes:**
   - Keep your `.env` file secure and never commit it to version control
   - The redirect URL must match exactly what you configured in Atlassian
   - Each user will need to authorize your app with their own Confluence account

## HTTPS Setup for Slack

Slack requires HTTPS redirect URLs, so we need to run the backend with HTTPS:

1. **Set USE_HTTPS=true in your `.env` file**
2. **The backend will automatically generate a self-signed certificate**
3. **When you first visit https://localhost:8085, your browser will show a security warning**
4. **Click "Advanced" → "Proceed to localhost (unsafe)" to accept the self-signed certificate**

## Testing the OAuth Flow

1. **Start the backend with HTTPS:**
   ```bash
   cd backend
   USE_HTTPS=true go run main.go
   ```

2. **Start the frontend:**
   ```bash
   cd frontend
   npm start
   ```

3. **Test the connections:**
   - Open http://localhost:3000 (frontend)
   - First visit https://localhost:8085 and accept the certificate warning
   - Expand the "Data Sources" panel
   - Click "Connect to [Service]" buttons
   - You should be redirected for authorization

## Troubleshooting

- **"Client not found" error**: Check your Client ID is correct
- **"Invalid redirect URI" error**: Ensure the callback URL matches exactly
- **Permission errors**: Verify you added the correct scopes
- **CORS errors**: Make sure both frontend and backend are running