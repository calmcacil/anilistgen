package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSeasonJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{
		{TVDBID: 12345, Title: "Test Show"},
		{TVDBID: 67890, Title: "Another Show"},
	}

	err := WriteSeasonJSON(dir, "WINTER", 2026, shows)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	path := filepath.Join(dir, "winter-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var so SeasonOutput
	if err := json.Unmarshal(data, &so); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if so.Season != "winter" {
		t.Errorf("expected season 'winter', got %q", so.Season)
	}
	if so.Year != 2026 {
		t.Errorf("expected year 2026, got %d", so.Year)
	}
	if len(so.Shows) != 2 {
		t.Fatalf("expected 2 shows, got %d", len(so.Shows))
	}
	if so.Shows[0].TVDBID != 12345 {
		t.Errorf("expected TVDB 12345, got %d", so.Shows[0].TVDBID)
	}
	if so.Shows[0].Title != "Test Show" {
		t.Errorf("expected title 'Test Show', got %q", so.Shows[0].Title)
	}
}

func TestWriteSeasonJSON_Compact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{{TVDBID: 1, Title: "T"}}

	err := WriteSeasonJSON(dir, "spring", 2026, shows)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	path := filepath.Join(dir, "spring-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if strings.Contains(string(data), "\n") {
		t.Error("expected compact JSON (no newlines)")
	}
}

func TestWriteYearJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{
		{TVDBID: 1, Title: "A"},
		{TVDBID: 2, Title: "B"},
	}

	err := WriteYearJSON(dir, 2026, shows)
	if err != nil {
		t.Fatalf("WriteYearJSON: %v", err)
	}

	path := filepath.Join(dir, "2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var yo YearOutput
	if err := json.Unmarshal(data, &yo); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if yo.Year != 2026 {
		t.Errorf("expected year 2026, got %d", yo.Year)
	}
	if len(yo.Shows) != 2 {
		t.Errorf("expected 2 shows, got %d", len(yo.Shows))
	}
}

func TestWriteAllJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	seasonal := map[string][]Show{
		"WINTER-2026": {{TVDBID: 1, Title: "Winter Show"}},
		"SPRING-2026": {{TVDBID: 2, Title: "Spring Show"}},
	}

	err := WriteAllJSON(dir, seasonal)
	if err != nil {
		t.Fatalf("WriteAllJSON: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 files (2 seasonal + 1 yearly), got %d", len(entries))
	}

	files := map[string]bool{}
	for _, e := range entries {
		files[e.Name()] = true
	}

	if !files["winter-2026.json"] {
		t.Error("missing winter-2026.json")
	}
	if !files["spring-2026.json"] {
		t.Error("missing spring-2026.json")
	}
	if !files["2026.json"] {
		t.Error("missing 2026.json")
	}
}
