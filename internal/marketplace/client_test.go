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
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestV2SearchFilter(t *testing.T) {
	public := true

	cases := map[string]struct {
		params SearchParams
		want   string
	}{
		"Query": {
			params: SearchParams{Query: "provider-aws-s3"},
			want:   "query = 'provider-aws-s3'",
		},
		"AllFiltersAreANDed": {
			params: SearchParams{Query: "s3", AccountName: "upbound", PackageType: "provider", Tier: "Official", Public: &public},
			want:   "query = 's3' AND packageType = 'Provider' AND account = 'upbound' AND tier = 'official' AND public = true",
		},
		"QuoteIsEscaped": {
			params: SearchParams{Query: "it's"},
			want:   `query = 'it\'s'`,
		},
		"Empty": {
			params: SearchParams{},
			want:   "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := v2SearchFilter(tc.params); got != tc.want {
				t.Errorf("v2SearchFilter(): want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSearchPackagesRequest(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		if _, err := w.Write([]byte(`{"packages":[{"account":"upbound","repository":"provider-aws-s3","packageType":"Provider","downloadCount":42}],"count":1,"size":20,"page":0}`)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient()
	c.SetBaseURL(srv.URL)

	resp, err := c.SearchPackages(context.Background(), SearchParams{Query: "provider-aws-s3", Size: 900})
	if err != nil {
		t.Fatalf("SearchPackages(): unexpected error: %v", err)
	}

	if got := gotQuery["filter"]; len(got) != 1 || got[0] != "query = 'provider-aws-s3'" {
		t.Errorf("filter: want a single %q, got %v", "query = 'provider-aws-s3'", got)
	}
	if got := gotQuery.Get("size"); got != "500" {
		t.Errorf("size: want the request to be capped at 500, got %s", got)
	}
	if resp.Count != 1 {
		t.Errorf("Count: want 1, got %d", resp.Count)
	}
	if len(resp.Packages) != 1 || resp.Packages[0].PackageType != "Provider" || resp.Packages[0].DownloadCount != 42 {
		t.Errorf("Packages: want the package fields decoded, got %+v", resp.Packages)
	}
}

func TestPackageVersionUnmarshalJSON(t *testing.T) {
	var versions []PackageVersion
	if err := json.Unmarshal([]byte(`["v2.7.1",{"display_versions":"v2.7.0","digest":"sha256:abc"}]`), &versions); err != nil {
		t.Fatalf("Unmarshal(): unexpected error: %v", err)
	}

	if len(versions) != 2 || versions[0].Version != "v2.7.1" || versions[1].Version != "v2.7.0" || versions[1].Digest != "sha256:abc" {
		t.Errorf("want both the v1 string and the v2 object form decoded, got %+v", versions)
	}
}
