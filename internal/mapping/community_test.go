package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommunityMapping_Lookup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "tvdb-mal.yaml")
	content := []byte(`AnimeMap:
  - malid: 16498
    tvdbid: 12345
    title: "Test Show"
  - malid: 30230
    tvdbid: 67890
    title: "Another Show"
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := LoadCommunityMapping(path)
	if err != nil {
		t.Fatalf("LoadCommunityMapping: %v", err)
	}

	tvdb, ok := cm.Lookup(16498)
	if !ok {
		t.Error("expected to find MAL 16498")
	}
	if tvdb != 12345 {
		t.Errorf("expected TVDB 12345, got %d", tvdb)
	}

	tvdb, ok = cm.Lookup(30230)
	if !ok {
		t.Error("expected to find MAL 30230")
	}
	if tvdb != 67890 {
		t.Errorf("expected TVDB 67890, got %d", tvdb)
	}
}

func TestCommunityMapping_Lookup_Missing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "tvdb-mal.yaml")
	content := []byte(`AnimeMap:
  - malid: 16498
    tvdbid: 12345
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := LoadCommunityMapping(path)
	if err != nil {
		t.Fatalf("LoadCommunityMapping: %v", err)
	}

	_, ok := cm.Lookup(99999)
	if ok {
		t.Error("expected false for missing MAL ID")
	}
}

func TestCommunityMapping_SkipsInvalidEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "tvdb-mal.yaml")
	content := []byte(`AnimeMap:
  - malid: 0
    tvdbid: 12345
  - malid: 16498
    tvdbid: 0
  - malid: 30230
    tvdbid: 67890
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := LoadCommunityMapping(path)
	if err != nil {
		t.Fatalf("LoadCommunityMapping: %v", err)
	}

	if len(cm.data) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(cm.data))
	}
}
