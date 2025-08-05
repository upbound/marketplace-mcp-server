// /*
// Copyright 2025 The Upbound Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

/*
Package auth handles authenticating with Upbound.
*/
package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UPConfig represents the UP CLI configuration structure.
type UPConfig struct {
	Upbound struct {
		Default  string             `json:"default"`
		Profiles map[string]Profile `json:"profiles"`
	} `json:"upbound"`
}

// Profile represents a UP CLI profile.
type Profile struct {
	ID           string `json:"id"`
	ProfileType  string `json:"profileType"`
	Type         string `json:"type"`
	Session      string `json:"session"`
	Account      string `json:"account,omitempty"`
	Organization string `json:"organization"`
	Domain       string `json:"domain"`
}

// Token represents an authentication token.
type Token struct {
	AccessToken string `json:"access_token"` //nolint:tagliatelle // External API.
	TokenType   string `json:"token_type"`   //nolint:tagliatelle // External API.
}

// Manager handles UP CLI configuration reading.
type Manager struct {
	configPath string
}

// NewManager creates a new authentication manager.
func NewManager() *Manager {
	var configPath string

	// Check if config path is specified via environment variable
	if envPath := os.Getenv("UP_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	} else {
		// Check for mounted config first (Docker container scenario)
		mountedPath := "/mcp/.up/config.json"
		if _, err := os.Stat(mountedPath); err == nil {
			configPath = mountedPath
		} else {
			// Fallback to user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				// Fallback to environment variable if available
				homeDir = os.Getenv("HOME")
			}
			configPath = filepath.Join(homeDir, ".up", "config.json")
		}
	}

	return &Manager{
		configPath: configPath,
	}
}

// GetCurrentToken returns the session token from the current UP CLI profile.
func (m *Manager) GetCurrentToken() (*Token, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load UP CLI config: %w", err)
	}

	// Get the default profile name
	defaultProfile := config.Upbound.Default
	if defaultProfile == "" {
		return nil, fmt.Errorf("no default profile set in UP CLI config")
	}

	// Get the profile
	profile, exists := config.Upbound.Profiles[defaultProfile]
	if !exists {
		return nil, fmt.Errorf("default profile '%s' not found in UP CLI config", defaultProfile)
	}

	// Check if session token exists
	if profile.Session == "" {
		return nil, fmt.Errorf("no session token found in profile '%s'. Please run 'up login' to authenticate", defaultProfile)
	}

	return &Token{
		AccessToken: profile.Session,
		TokenType:   "Session",
	}, nil
}

// GetTokenForProfile returns the session token for a specific profile.
func (m *Manager) GetTokenForProfile(profileName string) (*Token, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load UP CLI config: %w", err)
	}

	// Get the profile
	profile, exists := config.Upbound.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found in UP CLI config", profileName)
	}

	// Check if session token exists
	if profile.Session == "" {
		return nil, fmt.Errorf("no session token found in profile '%s'. Please run 'up login' to authenticate", profileName)
	}

	return &Token{
		AccessToken: profile.Session,
		TokenType:   "Session",
	}, nil
}

// GetCurrentProfile returns the current profile information.
func (m *Manager) GetCurrentProfile() (*Profile, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load UP CLI config: %w", err)
	}

	// Get the default profile name
	defaultProfile := config.Upbound.Default
	if defaultProfile == "" {
		return nil, fmt.Errorf("no default profile set in UP CLI config")
	}

	// Get the profile
	profile, exists := config.Upbound.Profiles[defaultProfile]
	if !exists {
		return nil, fmt.Errorf("default profile '%s' not found in UP CLI config", defaultProfile)
	}

	return &profile, nil
}

// ListProfiles returns all available profiles.
func (m *Manager) ListProfiles() (map[string]Profile, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load UP CLI config: %w", err)
	}

	return config.Upbound.Profiles, nil
}

// GetDefaultProfileName returns the name of the default profile.
func (m *Manager) GetDefaultProfileName() (string, error) {
	config, err := m.loadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load UP CLI config: %w", err)
	}

	if config.Upbound.Default == "" {
		return "", fmt.Errorf("no default profile set in UP CLI config")
	}

	return config.Upbound.Default, nil
}

// GetCurrentServerURL returns the server URL from the current profile.
func (m *Manager) GetCurrentServerURL() (string, error) {
	profile, err := m.GetCurrentProfile()
	if err != nil {
		return "", fmt.Errorf("failed to get current profile: %w", err)
	}

	if profile.Domain == "" {
		// Default to api.upbound.io if no domain is specified
		return "https://api.upbound.io", nil
	}

	// Parse the domain as a URL
	u, err := url.Parse(profile.Domain)
	if err != nil {
		return "", fmt.Errorf("failed to parse domain URL: %w", err)
	}

	// Add api. subdomain if not already present
	if !strings.HasPrefix(u.Host, "api.") {
		u.Host = "api." + u.Host
	}

	return u.String(), nil
}

// ValidateToken checks if the current token is valid (non-empty).
func (m *Manager) ValidateToken() error {
	token, err := m.GetCurrentToken()
	if err != nil {
		return err
	}

	if token.AccessToken == "" {
		return fmt.Errorf("empty session token")
	}

	return nil
}

// Legacy methods for backward compatibility - these now return errors
// since we don't do interactive authentication anymore

// Login is deprecated - use UP CLI authentication instead.
func (m *Manager) Login(_ any) (*Token, error) {
	return nil, fmt.Errorf("interactive login not supported. Please use 'up login' to authenticate with UP CLI")
}

// GetToken is deprecated - use GetCurrentToken instead.
func (m *Manager) GetToken() *Token {
	token, err := m.GetCurrentToken()
	if err != nil {
		return nil
	}
	return token
}

// RefreshToken is not applicable for session-based auth.
func (m *Manager) RefreshToken(_ any) (*Token, error) {
	return nil, fmt.Errorf("session token refresh not supported, please run 'up login' to re-authenticate")
}

// loadConfig loads the UP CLI configuration from disk.
func (m *Manager) loadConfig() (*UPConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("UP CLI config not found at %s. Please run 'up login' first", m.configPath)
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read UP CLI config: %w", err)
	}

	// Parse JSON
	var config UPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse UP CLI config: %w", err)
	}

	return &config, nil
}
