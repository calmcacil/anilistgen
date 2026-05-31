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

	err := WriteSeasonJSON(dir, "series", "WINTER", 2026, shows)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	path := filepath.Join(dir, "series-winter-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var got []Show
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 shows, got %d", len(got))
	}
	if got[0].TVDBID != 12345 {
		t.Errorf("expected TVDB 12345, got %d", got[0].TVDBID)
	}
	if got[0].Title != "Test Show" {
		t.Errorf("expected title 'Test Show', got %q", got[0].Title)
	}
}

func TestWriteSeasonJSON_Compact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{{TVDBID: 1, Title: "T"}}

	err := WriteSeasonJSON(dir, "series", "spring", 2026, shows)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	path := filepath.Join(dir, "series-spring-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if strings.Contains(string(data), "\n") {
		t.Error("expected compact JSON (no newlines)")
	}
}

func TestWriteSeasonJSON_SkipsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := WriteSeasonJSON(dir, "movies", "WINTER", 2026, nil)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	_, err = os.Stat(filepath.Join(dir, "movies-winter-2026.json"))
	if !os.IsNotExist(err) {
		t.Error("expected no file for empty shows")
	}
}

func TestWriteYearJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{
		{TVDBID: 1, Title: "A"},
		{TVDBID: 2, Title: "B"},
	}

	err := WriteYearJSON(dir, "series", 2026, shows)
	if err != nil {
		t.Fatalf("WriteYearJSON: %v", err)
	}

	path := filepath.Join(dir, "series-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var got []Show
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("expected 2 shows, got %d", len(got))
	}
}

func TestWriteAllJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	seasonal := map[string][]Show{
		"WINTER-2026": {{TVDBID: 1, Title: "Winter Show"}},
		"SPRING-2026": {{TVDBID: 2, Title: "Spring Show"}},
	}

	err := WriteAllJSON(dir, "series", seasonal)
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

	if !files["series-winter-2026.json"] {
		t.Error("missing series-winter-2026.json")
	}
	if !files["series-spring-2026.json"] {
		t.Error("missing series-spring-2026.json")
	}
	if !files["series-2026.json"] {
		t.Error("missing series-2026.json")
	}
}

func TestWriteSeasonJSON_StartsAsArray(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shows := []Show{{TVDBID: 1, Title: "T"}}

	err := WriteSeasonJSON(dir, "movies", "fall", 2026, shows)
	if err != nil {
		t.Fatalf("WriteSeasonJSON: %v", err)
	}

	path := filepath.Join(dir, "movies-fall-2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if data[0] != '[' {
		t.Errorf("expected JSON array starting with '[', got %c", data[0])
	}
}

func TestWriteAllJSON_MultipleCategories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	series := map[string][]Show{
		"WINTER-2026": {{TVDBID: 1, Title: "Series A"}},
	}
	movies := map[string][]Show{
		"WINTER-2026": {{TVDBID: 2, Title: "Movie A"}},
	}

	if err := WriteAllJSON(dir, "series", series); err != nil {
		t.Fatal(err)
	}
	if err := WriteAllJSON(dir, "movies", movies); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}

	if !names["series-winter-2026.json"] {
		t.Error("missing series-winter-2026.json")
	}
	if !names["movies-winter-2026.json"] {
		t.Error("missing movies-winter-2026.json")
	}
	if !names["series-2026.json"] {
		t.Error("missing series-2026.json")
	}
	if !names["movies-2026.json"] {
		t.Error("missing movies-2026.json")
	}
}
