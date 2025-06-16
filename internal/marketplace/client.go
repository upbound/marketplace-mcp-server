package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL = "https://registry.upbound.io"
	UserAgent      = "marketplace-mcp-server/1.0"
)

// Client represents a marketplace API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient creates a new marketplace client
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.Token = token
}

// SearchPackages searches for packages using v1 or v2 API
func (c *Client) SearchPackages(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	endpoint := "/v2/search"
	if params.UseV1 {
		endpoint = "/v1/search"
	}

	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if params.Query != "" {
		if params.UseV1 {
			q.Set("query", params.Query)
		} else {
			// V2 uses filter parameter with AIP-160 format
			q.Set("filter", fmt.Sprintf("query = '%s'", params.Query))
		}
	}
	if params.Family != "" {
		if params.UseV1 {
			q.Set("family", params.Family)
		} else {
			q.Set("filter", fmt.Sprintf("family = '%s'", params.Family))
		}
	}
	if params.PackageType != "" {
		if params.UseV1 {
			q.Set("packageType", params.PackageType)
		} else {
			q.Set("filter", fmt.Sprintf("packageType = '%s'", params.PackageType))
		}
	}
	if params.AccountName != "" {
		if params.UseV1 {
			q.Set("accountName", params.AccountName)
		} else {
			q.Set("filter", fmt.Sprintf("accountName = '%s'", params.AccountName))
		}
	}
	if params.Size > 0 {
		q.Set("size", fmt.Sprintf("%d", params.Size))
	}
	if params.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.Public != nil {
		q.Set("public", fmt.Sprintf("%t", *params.Public))
	}
	if params.Tier != "" {
		if params.UseV1 {
			q.Set("tier", params.Tier)
		} else {
			q.Set("filter", fmt.Sprintf("tier = '%s'", params.Tier))
		}
	}
	if params.Starred != nil && *params.Starred {
		q.Set("starred", "true")
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// GetPackageMetadata gets metadata for a specific package
func (c *Client) GetPackageMetadata(ctx context.Context, account, repo, version string, useV1 bool) (*PackageMetadata, error) {
	endpoint := fmt.Sprintf("/v2/packageMetadata/%s/%s", account, repo)
	if version != "" {
		endpoint = fmt.Sprintf("/v1/packageMetadata/%s/%s/%s", account, repo, version)
		useV1 = true
	}
	if useV1 && version == "" {
		endpoint = fmt.Sprintf("/v1/packageMetadata/%s/%s", account, repo)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var metadata PackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &metadata, nil
}

// GetPackageAssets gets assets for a specific package version
func (c *Client) GetPackageAssets(ctx context.Context, account, repo, version, assetType string) (*AssetResponse, error) {
	endpoint := fmt.Sprintf("/v2/packages/%s/%s/%s/assets", account, repo, version)

	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("type", assetType)
	q.Set("redirect", "false") // Get the URL instead of redirecting
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTemporaryRedirect {
		location := resp.Header.Get("Location")
		return &AssetResponse{URL: location}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var assetResp AssetResponse
	if err := json.NewDecoder(resp.Body).Decode(&assetResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &assetResp, nil
}

// GetRepositories gets repositories for an account
func (c *Client) GetRepositories(ctx context.Context, account string, params RepositoryParams) (*RepositoryResponse, error) {
	endpoint := fmt.Sprintf("/v2/repositories/%s", account)
	if params.UseV1 {
		endpoint = fmt.Sprintf("/v1/repositories/%s", account)
	}

	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if params.Size > 0 {
		q.Set("size", fmt.Sprintf("%d", params.Size))
	}
	if params.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.Filter != "" && !params.UseV1 {
		q.Set("filter", params.Filter)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var repoResp RepositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repoResp, nil
}
