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

package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

const (
	userAgent = "marketplace-mcp-server/1.0"
)

// Client represents a marketplace API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string

	log logging.Logger
}

// Option enables overriding the underlying Client.
type Option func(*Client)

// WithLogger overrides the default logger for the Client.
func WithLogger(log logging.Logger) Option {
	return func(c *Client) {
		c.log = log
	}
}

// NewClient creates a new marketplace client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		BaseURL: "", // Will be set by the server from UP CLI profile
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: logging.NewNopLogger(),
	}

	for _, o := range opts {
		o(c)
	}
	return c
}

// SetToken sets the authentication token.
func (c *Client) SetToken(token string) {
	c.Token = token
}

// SetBaseURL sets the base URL for the marketplace API.
func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

// SearchPackages searches for packages using v1 or v2 API.
func (c *Client) SearchPackages(ctx context.Context, params SearchParams) (*SearchResponse, error) { //nolint:gocognit // This method is unfortunately above our complexity level. Be wary of increasing its complexity.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

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

// GetPackageMetadata gets metadata for a specific package.
func (c *Client) GetPackageMetadata(ctx context.Context, account, repo, version string, useV1 bool) (*PackageMetadata, error) {
	endpoint := fmt.Sprintf("/v2/packageMetadata/%s/%s", account, repo)
	if version != "" {
		endpoint = fmt.Sprintf("/v1/packageMetadata/%s/%s/%s", account, repo, version)
		useV1 = true
	}
	if useV1 && version == "" {
		endpoint = fmt.Sprintf("/v1/packageMetadata/%s/%s", account, repo)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

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

// GetPackageAssets gets assets for a specific package version.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusTemporaryRedirect {
		location := resp.Header.Get("Location")
		return &AssetResponse{URL: location}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Try to decode as single object first
	var assetResp AssetResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try single object first
	if err := json.Unmarshal(body, &assetResp); err != nil {
		// Try array format
		var assetArray []AssetResponse
		if err := json.Unmarshal(body, &assetArray); err != nil {
			return nil, fmt.Errorf("failed to decode response as object or array: %w", err)
		}
		// Return first item if array
		if len(assetArray) > 0 {
			return &assetArray[0], nil
		}
		return &AssetResponse{}, nil
	}

	return &assetResp, nil
}

// GetRepositories gets repositories for an account.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the raw response to debug structure
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var repoResp RepositoryResponse
	if err := json.Unmarshal(body, &repoResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Fill in missing account information for repositories
	for i := range repoResp.Repositories {
		if repoResp.Repositories[i].Account == "" {
			repoResp.Repositories[i].Account = account
		}
	}

	return &repoResp, nil
}

// GetV1PackagesAccountRepositoryVersionResources - [/v1/packages/{account}/{repositoryName}/{version}/resources].
func (c *Client) GetV1PackagesAccountRepositoryVersionResources(ctx context.Context, account, repositoryName, version string) (*PackageResources, error) {
	endpoint := fmt.Sprintf("/v1/packages/%s/%s/%s/resources", account, repositoryName, version)

	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the raw response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var res PackageResources
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &res, nil
}

// GetV1PackagesAccountRepositoryVersionResourcesGroupKindComposition - [/v1/packages/{account}/{repositoryName}/{version}/resources/{resourceGroup}/{resourceKind}/compositions/{compositionName}].
func (c *Client) GetV1PackagesAccountRepositoryVersionResourcesGroupKindComposition(ctx context.Context, account, repositoryName, version, resourceGroup, resourceKind, compositionName string) (string, error) {
	endpoint := fmt.Sprintf("/v1/packages/%s/%s/%s/resources/%s/%s/compositions/%s", account, repositoryName, version, resourceGroup, resourceKind, compositionName)
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the raw response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// GetV1PackagesAccountRepositoryVersionResourcesGroupKind - [/v1/packages/{account}/{repositoryName}/{version}/resources/{resourceGroup}/{resourceKind}].
func (c *Client) GetV1PackagesAccountRepositoryVersionResourcesGroupKind(ctx context.Context, account, repositoryName, version, resourceGroup, resourceKind string) (string, error) {
	endpoint := fmt.Sprintf("/v1/packages/%s/%s/%s/resources/%s/%s", account, repositoryName, version, resourceGroup, resourceKind)
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the raw response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// GetV1PackagesAccountRepositoryVersionResourcesGroupKindExamples - [/v1/packages/{account}/{repositoryName}/{version}/resources/{resourceGroup}/{resourceKind}/examples].
func (c *Client) GetV1PackagesAccountRepositoryVersionResourcesGroupKindExamples(ctx context.Context, account, repositoryName, version, resourceGroup, resourceKind string) (*Examples, error) {
	endpoint := fmt.Sprintf("/v1/packages/%s/%s/%s/resources/%s/%s/examples", account, repositoryName, version, resourceGroup, resourceKind)
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		// Use session cookie authentication like UP CLI
		req.AddCookie(&http.Cookie{
			Name:  "SID",
			Value: c.Token,
		})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Info("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication required for this endpoint")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the raw response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ex Examples
	if err := json.Unmarshal(body, &ex); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ex, nil
}
