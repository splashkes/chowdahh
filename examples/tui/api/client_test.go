package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCategoriesUsesStreamDiscoveryEndpoint(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		json.NewEncoder(w).Encode(Envelope[CategoriesData]{
			Data: CategoriesData{
				Streams: []Category{{Slug: "latest", Label: "Latest"}},
				Count:   1,
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	env, err := client.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories returned error: %v", err)
	}
	if requestedPath != "/api/v1/streams" {
		t.Fatalf("GetCategories requested %q, want /api/v1/streams", requestedPath)
	}
	if len(env.Data.Streams) != 1 || env.Data.Streams[0].Slug != "latest" {
		t.Fatalf("decoded streams = %#v", env.Data.Streams)
	}
}

func TestSearchDecodesLiveResultsShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Envelope[SearchResult]{
			Data: SearchResult{
				Query:   "NASA",
				Results: []Card{{ID: "card-1", Headline: "NASA story"}},
				Count:   1,
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	env, err := client.Search("NASA", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(env.Data.Results) != 1 || env.Data.Results[0].Headline != "NASA story" {
		t.Fatalf("decoded results = %#v", env.Data.Results)
	}
}
