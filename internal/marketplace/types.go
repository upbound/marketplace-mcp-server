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

import "time"

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

// SearchResponse represents the response from search endpoints.
type SearchResponse struct {
	Packages []Package `json:"packages,omitempty"`
	Total    int       `json:"total,omitempty"`
	Page     int       `json:"page,omitempty"`
	Size     int       `json:"size,omitempty"`
}

// Package represents a package in search results.
type Package struct {
	Account     string         `json:"account"`
	Repository  string         `json:"repository"`
	Name        string         `json:"name"`
	Version     string         `json:"version,omitempty"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type,omitempty"`
	Public      bool           `json:"public"`
	Tier        string         `json:"tier,omitempty"`
	Stars       int            `json:"stars,omitempty"`
	Downloads   int            `json:"downloads,omitempty"`
	CreatedAt   time.Time      `json:"createdAt,omitempty"`
	UpdatedAt   time.Time      `json:"updatedAt,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Keywords    []string       `json:"keywords,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// PackageMetadata represents detailed package metadata.
type PackageMetadata struct {
	Account       string         `json:"account"`
	Repository    string         `json:"repository"`
	Name          string         `json:"name"`
	Version       string         `json:"version,omitempty"`
	Description   string         `json:"description,omitempty"`
	Type          string         `json:"type,omitempty"`
	Public        bool           `json:"public"`
	Tier          string         `json:"tier,omitempty"`
	Stars         int            `json:"stars,omitempty"`
	Downloads     int            `json:"downloads,omitempty"`
	CreatedAt     time.Time      `json:"createdAt,omitempty"`
	UpdatedAt     time.Time      `json:"updatedAt,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Keywords      []string       `json:"keywords,omitempty"`
	Versions      []string       `json:"versions,omitempty"`
	LatestVersion string         `json:"latestVersion,omitempty"`
	Documentation string         `json:"documentation,omitempty"`
	Homepage      string         `json:"homepage,omitempty"`
	License       string         `json:"license,omitempty"`
	Dependencies  []Dependency   `json:"dependencies,omitempty"`
	CRDs          []CRD          `json:"crds,omitempty"`
	Examples      []Example      `json:"examples,omitempty"`
	Compositions  []Composition  `json:"compositions,omitempty"`
	Functions     []Function     `json:"functions,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
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
