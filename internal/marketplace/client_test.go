package marketplace

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.BaseURL != "" {
		t.Errorf("Expected BaseURL to be empty initially, got %s", client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestSetToken(t *testing.T) {
	client := NewClient()
	token := "test-token"

	client.SetToken(token)

	if client.Token != token {
		t.Errorf("Expected token to be %s, got %s", token, client.Token)
	}
}
