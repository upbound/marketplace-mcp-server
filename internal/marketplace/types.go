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
	"encoding/json"
	"time"
)

// SearchParams represents search parameters for package search.
type SearchParams struct {
	Query       string
	Family      string
	PackageType string
	AccountName string
	Size        int
	Page        int
	Public      *bool
	Tier        string
	Starred     *bool
	Type        string
	UseV1       bool
}

// RepositoryParams represents parameters for repository queries.
type RepositoryParams struct {
	Size   int
	Page   int
	Filter string
	UseV1  bool
}

// SearchResponse represents the response from search endpoints. Both the v1
// and the v2 search API report the number of matches as "count".
type SearchResponse struct {
	Packages []Package `json:"packages,omitempty"`
	Count    int       `json:"count,omitempty"`
	Page     int       `json:"page,omitempty"`
	Size     int       `json:"size,omitempty"`
}

// Package represents a package in search results.
type Package struct {
	Account       string              `json:"account"`
	Repository    string              `json:"repository"`
	RepoKey       string              `json:"repoKey,omitempty"`
	Name          string              `json:"name"`
	Version       string              `json:"version,omitempty"`
	Description   string              `json:"description,omitempty"`
	PackageType   string              `json:"packageType,omitempty"`
	Public        bool                `json:"public"`
	Tier          string              `json:"tier,omitempty"`
	PkgDigest     string              `json:"pkgDigest,omitempty"`
	FamilyRepoKey string              `json:"familyRepoKey,omitempty"`
	DownloadCount int                 `json:"downloadCount,omitempty"`
	IconURL       string              `json:"iconURL,omitempty"`
	CreatedAt     time.Time           `json:"createdAt,omitempty"`
	UpdatedAt     time.Time           `json:"updatedAt,omitempty"`
	Annotations   map[string][]string `json:"annotations,omitempty"`
	Highlights    map[string][]string `json:"highlight,omitempty"`
}

// PackageMetadata represents detailed package metadata. The versioned
// endpoint returns a subset of these fields plus the version specific ones.
type PackageMetadata struct {
	RepoKey           string             `json:"repoKey"`
	Type              string             `json:"type,omitempty"`
	Tier              string             `json:"tier,omitempty"`
	Public            bool               `json:"public"`
	CurrentVersion    string             `json:"currentVersion,omitempty"`
	FamilyRepoKey     string             `json:"familyRepoKey,omitempty"`
	Versions          []PackageVersion   `json:"versions,omitempty"`
	RelatedRepository *RelatedRepository `json:"relatedRepository,omitempty"`

	// Fields only returned when a specific version is requested.
	Digest         string `json:"digest,omitempty"`
	PublishedAt    string `json:"publishedAt,omitempty"`
	EndOfSupport   string `json:"endOfSupport,omitempty"`
	EndOfLife      string `json:"endOfLife,omitempty"`
	HasSignature   bool   `json:"hasSignature,omitempty"`
	HasAttestation bool   `json:"hasAttestation,omitempty"`
	Languages      string `json:"languages,omitempty"`
}

// RelatedRepository represents a repository related to the requested one, for
// example the long term support variant of a package.
type RelatedRepository struct {
	RepoKey          string           `json:"repoKey"`
	Public           bool             `json:"public"`
	SubscriptionTier string           `json:"subscriptionTier,omitempty"`
	Versions         []PackageVersion `json:"versions,omitempty"`
}

// PackageVersion represents a published version of a package. The v2
// packageMetadata endpoint returns objects while v1 returns plain version
// strings, so both encodings are accepted.
type PackageVersion struct {
	Version             string    `json:"display_versions,omitempty"` //nolint:tagliatelle // This is marshalling an external API.
	Digest              string    `json:"digest,omitempty"`
	FreeVersion         string    `json:"free_version,omitempty"`         //nolint:tagliatelle // This is marshalling an external API.
	SubscriptionVersion string    `json:"subscription_version,omitempty"` //nolint:tagliatelle // This is marshalling an external API.
	UpdatedAt           time.Time `json:"updated_at,omitempty"`           //nolint:tagliatelle // This is marshalling an external API.
}

// UnmarshalJSON decodes a package version from either a version string (v1) or
// a version object (v2).
func (v *PackageVersion) UnmarshalJSON(data []byte) error {
	var version string
	if err := json.Unmarshal(data, &version); err == nil {
		*v = PackageVersion{Version: version}
		return nil
	}

	type packageVersion PackageVersion // Avoid recursing into this method.
	var decoded packageVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*v = PackageVersion(decoded)

	return nil
}

// Dependency represents a package dependency.
type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Constraints string `json:"constraints,omitempty"`
}

// CRD represents a Custom Resource Definition.
type CRD struct {
	Name        string         `json:"name"`
	Group       string         `json:"group"`
	Version     string         `json:"version"`
	Kind        string         `json:"kind"`
	Plural      string         `json:"plural"`
	Singular    string         `json:"singular"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

// Example represents a usage example.
type Example struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content"`
	Type        string `json:"type,omitempty"` // yaml, json, etc.
}

// Composition represents a Crossplane composition.
type Composition struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Content     string                `json:"content"`
	Resources   []CompositionResource `json:"resources,omitempty"`
	Metadata    map[string]any        `json:"metadata,omitempty"`
}

// CompositionResource represents a resource in a composition.
type CompositionResource struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	Base map[string]any `json:"base,omitempty"`
}

// Function represents a Crossplane function.
type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version"`
	Image       string         `json:"image"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// AssetResponse represents the response from asset endpoints.
type AssetResponse struct {
	URL     string `json:"url,omitempty"`
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
}

// RepositoryResponse represents the response from repository endpoints.
type RepositoryResponse struct {
	Repositories []Repository `json:"repositories,omitempty"`
	Count        int          `json:"count,omitempty"`
	Page         int          `json:"page,omitempty"`
	Size         int          `json:"size,omitempty"`
}

// Repository represents a repository.
type Repository struct {
	Account      string    `json:"account"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Type         string    `json:"type,omitempty"`
	Public       bool      `json:"public"`
	Policy       string    `json:"policy,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	PackageCount int       `json:"packageCount,omitempty"`
}

// AuthResponse represents authentication response.
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      User      `json:"user"`
}

// User represents a user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
}

// PackageType enumerates the valid types of Crossplane packages.
type PackageType string

// CRDMeta contains CustomResourceDefinition metadata.
type CRDMeta struct {
	Group          string   `json:"group"`
	Kind           string   `json:"kind"`
	Versions       []string `json:"versions"`
	StorageVersion string   `json:"storageVersion"`
	Scope          string   `json:"scope"`
}

// XRDMeta contains CompositeResourceDefinition metadata.
type XRDMeta struct {
	Group                string   `json:"group"`
	Kind                 string   `json:"kind"`
	Versions             []string `json:"versions"`
	ReferenceableVersion string   `json:"referenceableVersion"`
}

// CompositionMeta contains Composition metadata.
type CompositionMeta struct {
	Name          string `json:"name"`
	ResourceCount int    `json:"resourceCount"`
	XrdAPIVersion string `json:"xrdApiVersion"` //nolint:tagliatelle // This is marshalling an external API.
	XrdKind       string `json:"xrdKind"`
}

// PackageMeta contains package metadata.
type PackageMeta struct {
	Account       string              `json:"account"`
	Repository    string              `json:"repository"`
	RepoKey       string              `json:"repoKey"`
	Name          string              `json:"name"`
	PackageType   PackageType         `json:"packageType"`
	Public        bool                `json:"public"`
	Tier          string              `json:"tier"`
	PkgDigest     string              `json:"pkgDigest"`
	FamilyRepoKey *string             `json:"familyRepoKey,omitempty"`
	FamilyCount   *uint               `json:"familyCount,omitempty"`
	Highlights    map[string][]string `json:"highlight,omitempty"`
}

// PackageResources contains extended package metadata that includes the
// resources in the package.
type PackageResources struct {
	PackageMeta `json:",inline"`

	CRDs         []CRDMeta         `json:"customResourceDefinitions"`
	XRDs         []XRDMeta         `json:"compositeResourceDefinitions"`
	Compositions []CompositionMeta `json:"compositions"`
}

// Examples return type for multiple examples.
type Examples struct {
	Examples []string `json:"examples"`
}
