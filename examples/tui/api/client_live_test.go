package api

import (
	"os"
	"testing"
)

func TestLiveDiscoveryAndSearchShapes(t *testing.T) {
	if os.Getenv("CHOWDAHH_LIVE_TEST") != "1" {
		t.Skip("set CHOWDAHH_LIVE_TEST=1 to run live API smoke test")
	}
	if testing.Short() {
		t.Skip("skipping live API smoke test in short mode")
	}

	client := NewClient("https://chowdahh.com", "")

	streams, err := client.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories via /api/v1/streams failed: %v", err)
	}
	if len(streams.Data.Categories) == 0 && len(streams.Data.Streams) == 0 {
		t.Fatalf("expected live stream discovery response to contain streams")
	}

	search, err := client.Search("NASA", 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if search.Data.Count > 0 && len(search.Data.Cards) == 0 && len(search.Data.Results) == 0 {
		t.Fatalf("expected live search response to contain cards or results")
	}
}
