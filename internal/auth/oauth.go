package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	// Upbound OAuth endpoints
	AuthURL  = "https://accounts.upbound.io/oauth/authorize"
	TokenURL = "https://accounts.upbound.io/oauth/token"

	// Local server for OAuth callback
	CallbackPort = "8765"
	CallbackPath = "/callback"
	RedirectURI  = "http://localhost:" + CallbackPort + CallbackPath
)

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// AuthManager handles OAuth authentication
type AuthManager struct {
	config *oauth2.Config
	server *http.Server
	token  *oauth2.Token
	done   chan bool
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(clientID, clientSecret string, scopes []string) *AuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  RedirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	return &AuthManager{
		config: config,
		done:   make(chan bool, 1),
	}
}

// Login initiates the OAuth login flow
func (am *AuthManager) Login(ctx context.Context) (*oauth2.Token, error) {
	// Generate state parameter for security
	state, err := generateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start local server for callback
	if err := am.startCallbackServer(state); err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer am.stopCallbackServer()

	// Generate authorization URL
	authURL := am.config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Open browser to authorization URL
	log.Printf("Opening browser for authentication: %s", authURL)
	if err := browser.OpenURL(authURL); err != nil {
		log.Printf("Failed to open browser automatically. Please visit: %s", authURL)
	}

	// Wait for callback or timeout
	select {
	case <-am.done:
		if am.token == nil {
			return nil, fmt.Errorf("authentication failed")
		}
		return am.token, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timeout: %w", ctx.Err())
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authentication timeout")
	}
}

// GetToken returns the current token
func (am *AuthManager) GetToken() *oauth2.Token {
	return am.token
}

// RefreshToken refreshes the OAuth token if needed
func (am *AuthManager) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	if am.token == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	tokenSource := am.config.TokenSource(ctx, am.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	am.token = newToken
	return newToken, nil
}

// startCallbackServer starts the local HTTP server for OAuth callback
func (am *AuthManager) startCallbackServer(expectedState string) error {
	mux := http.NewServeMux()

	mux.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		am.handleCallback(w, r, expectedState)
	})

	am.server = &http.Server{
		Addr:    ":" + CallbackPort,
		Handler: mux,
	}

	go func() {
		if err := am.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Callback server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// stopCallbackServer stops the local HTTP server
func (am *AuthManager) stopCallbackServer() {
	if am.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		am.server.Shutdown(ctx)
	}
}

// handleCallback handles the OAuth callback
func (am *AuthManager) handleCallback(w http.ResponseWriter, r *http.Request, expectedState string) {
	// Check for errors
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		http.Error(w, fmt.Sprintf("OAuth error: %s", errMsg), http.StatusBadRequest)
		am.done <- true
		return
	}

	// Verify state parameter
	state := r.URL.Query().Get("state")
	if state != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		am.done <- true
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		am.done <- true
		return
	}

	// Exchange code for token
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := am.config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		am.done <- true
		return
	}

	am.token = token

	// Send success response
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
		<head><title>Authentication Successful</title></head>
		<body>
			<h1>Authentication Successful!</h1>
			<p>You can now close this window and return to your application.</p>
			<script>window.close();</script>
		</body>
		</html>
	`))

	am.done <- true
}

// generateRandomState generates a random state parameter for OAuth
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthenticatedClient returns an HTTP client with OAuth token
func (am *AuthManager) GetAuthenticatedClient(ctx context.Context) *http.Client {
	if am.token == nil {
		return http.DefaultClient
	}
	return am.config.Client(ctx, am.token)
}
