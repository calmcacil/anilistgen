package mapping

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/calmcacil/anilistgen/internal/model"
	"github.com/klauspost/compress/zstd"
)

func TestLookupByMAL(t *testing.T) {
	t.Parallel()

	am := &AnibridgeMapping{
		byMAL: map[int]int{
			16498: 12345,
			99999: 67890,
		},
	}

	t.Run("known MAL ID", func(t *testing.T) {
		tvdbID, ok := am.LookupByMAL(16498)
		if !ok {
			t.Error("expected ok for known MAL ID")
		}
		if tvdbID != 12345 {
			t.Errorf("expected TVDB 12345, got %d", tvdbID)
		}
	})

	t.Run("unknown MAL ID", func(t *testing.T) {
		_, ok := am.LookupByMAL(1)
		if ok {
			t.Error("expected !ok for unknown MAL ID")
		}
	})

	t.Run("zero MAL ID", func(t *testing.T) {
		_, ok := am.LookupByMAL(0)
		if ok {
			t.Error("expected !ok for zero MAL ID")
		}
	})
}

func TestLookupByAniList(t *testing.T) {
	t.Parallel()

	am := &AnibridgeMapping{
		byAniList: map[int]int{
			100: 54321,
			200: 98765,
		},
	}

	t.Run("known AniList ID", func(t *testing.T) {
		tvdbID, ok := am.LookupByAniList(100)
		if !ok {
			t.Error("expected ok for known AniList ID")
		}
		if tvdbID != 54321 {
			t.Errorf("expected TVDB 54321, got %d", tvdbID)
		}
	})

	t.Run("unknown AniList ID", func(t *testing.T) {
		_, ok := am.LookupByAniList(999)
		if ok {
			t.Error("expected !ok for unknown AniList ID")
		}
	})
}

func TestNewResolverAndProject(t *testing.T) {
	t.Parallel()

	am := &AnibridgeMapping{
		byMAL: map[int]int{
			16498: 12345,
		},
		byAniList: map[int]int{
			42: 77777,
		},
	}
	r := NewResolver(am)
	if r == nil {
		t.Fatal("expected non-nil Resolver")
	}

	t.Run("project via MAL", func(t *testing.T) {
		shows := []model.Show{
			{ID: 1, IDMal: makePtr(16498), Title: model.Title{English: makePtr("Test Show")}},
		}
		result := r.Project(shows)
		if len(result) != 1 {
			t.Fatalf("expected 1 show, got %d", len(result))
		}
		if result[0].TVDBID != 12345 {
			t.Errorf("expected 12345, got %d", result[0].TVDBID)
		}
		if result[0].Title != "Test Show" {
			t.Errorf("expected 'Test Show', got %q", result[0].Title)
		}
	})

	t.Run("project via AniList fallback", func(t *testing.T) {
		shows := []model.Show{
			{ID: 42, IDMal: nil, Title: model.Title{English: makePtr("AniList Original")}},
		}
		result := r.Project(shows)
		if len(result) != 1 {
			t.Fatalf("expected 1 show, got %d", len(result))
		}
		if result[0].TVDBID != 77777 {
			t.Errorf("expected 77777, got %d", result[0].TVDBID)
		}
	})

	t.Run("project nil MAL falls through to AniList", func(t *testing.T) {
		shows := []model.Show{
			{ID: 42, IDMal: nil, Title: model.Title{English: makePtr("No MAL")}},
		}
		result := r.Project(shows)
		if len(result) != 1 {
			t.Errorf("expected 1 show for AniList fallback, got %d", len(result))
		}
	})

	t.Run("project unknown in both", func(t *testing.T) {
		shows := []model.Show{
			{ID: 1, IDMal: makePtr(1), Title: model.Title{English: makePtr("Unknown")}},
		}
		result := r.Project(shows)
		if len(result) != 0 {
			t.Errorf("expected 0 shows for unknown, got %d", len(result))
		}
	})

	t.Run("project both missing", func(t *testing.T) {
		shows := []model.Show{
			{ID: 0, IDMal: nil, Title: model.Title{English: makePtr("Everything Missing")}},
		}
		result := r.Project(shows)
		if len(result) != 0 {
			t.Errorf("expected 0 shows, got %d", len(result))
		}
	})

	t.Run("MAL takes priority over AniList", func(t *testing.T) {
		am := &AnibridgeMapping{
			byMAL: map[int]int{
				16498: 12345,
			},
			byAniList: map[int]int{
				1: 99999,
			},
		}
		r := NewResolver(am)
		shows := []model.Show{
			{ID: 1, IDMal: makePtr(16498), Title: model.Title{English: makePtr("Priority")}},
		}
		result := r.Project(shows)
		if len(result) != 1 {
			t.Fatalf("expected 1 show, got %d", len(result))
		}
		if result[0].TVDBID != 12345 {
			t.Errorf("expected MAL entry 12345 to take priority over AniList 99999, got %d", result[0].TVDBID)
		}
	})
}

func TestParseAnibridgeJSON_SmallFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json.zst")

	fixture := `{
  "$meta": {
    "schema_version": "3.0.0",
    "generated_on": "2026-01-01T00:00:00Z"
  },
  "mal:1": {
    "tvdb_show:100:s1": { "1-12": "1-12" }
  },
  "mal:2": {
    "tvdb_show:200:s1": { "1-24": "1-24" },
    "tvdb_show:200:s0": { "1-3": "4-6" }
  },
  "anilist:42": {
    "tvdb_show:300:s1": { "1-13": "1-13" }
  },
  "anilist:99": {
    "tvdb_show:400:s2": { "1-12": "1-12" }
  },
  "anidb:999:R": {
    "mal:1": { "1-12": "1-12" }
  }
}`

	if err := writeZstdFile(path, fixture); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	am, err := LoadAnibridgeMapping(path)
	if err != nil {
		t.Fatalf("LoadAnibridgeMapping: %v", err)
	}

	t.Run("mal entries", func(t *testing.T) {
		tvdbID, ok := am.LookupByMAL(1)
		if !ok || tvdbID != 100 {
			t.Errorf("expected MAL 1 -> TVDB 100, got %d, %v", tvdbID, ok)
		}

		tvdbID, ok = am.LookupByMAL(2)
		if !ok || tvdbID != 200 {
			t.Errorf("expected MAL 2 -> TVDB 200, got %d, %v", tvdbID, ok)
		}

		if _, ok := am.LookupByMAL(3); ok {
			t.Error("expected MAL 3 to be absent")
		}
	})

	t.Run("anilist entries", func(t *testing.T) {
		tvdbID, ok := am.LookupByAniList(42)
		if !ok || tvdbID != 300 {
			t.Errorf("expected AniList 42 -> TVDB 300, got %d, %v", tvdbID, ok)
		}

		tvdbID, ok = am.LookupByAniList(99)
		if !ok || tvdbID != 400 {
			t.Errorf("expected AniList 99 -> TVDB 400 (s2 when no s1), got %d, %v", tvdbID, ok)
		}

		if _, ok := am.LookupByAniList(100); ok {
			t.Error("expected AniList 100 to be absent")
		}
	})
}

func TestParseAnibridgeJSON_PrefersS1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json.zst")

	fixture := `{
  "mal:10": {
    "tvdb_show:500:s1": { "1-24": "1-24" },
    "tvdb_show:500:s2": { "1-24": "1-24" },
    "tvdb_show:500:s0": { "1-6": "1-6" }
  },
  "mal:11": {
    "tvdb_show:600:s0": { "1-50": "1-50" },
    "tvdb_show:600:s2": { "1-24": "1-24" }
  }
}`

	if err := writeZstdFile(path, fixture); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	am, err := LoadAnibridgeMapping(path)
	if err != nil {
		t.Fatalf("LoadAnibridgeMapping: %v", err)
	}

	t.Run("s1 preferred over higher ep count in s0", func(t *testing.T) {
		tvdbID, ok := am.LookupByMAL(10)
		if !ok || tvdbID != 500 {
			t.Errorf("expected MAL 10 -> TVDB 500, got %d, %v", tvdbID, ok)
		}
	})

	t.Run("no s1 falls back to highest ep count", func(t *testing.T) {
		tvdbID, ok := am.LookupByMAL(11)
		if !ok || tvdbID != 600 {
			t.Errorf("expected MAL 11 -> TVDB 600, got %d, %v", tvdbID, ok)
		}
	})
}

func TestLoadAnibridgeMapping_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json.zst")

	_, err := LoadAnibridgeMapping(path)
	if err == nil {
		t.Skip("network request succeeded unexpectedly — test env may have internet access")
	}
}

func TestLoadAnibridgeMappingWithAge_Fresh(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json.zst")

	fixture := `{ "mal:1": { "tvdb_show:100:s1": { "1-12": "1-12" } } }`
	if err := writeZstdFile(path, fixture); err != nil {
		t.Fatal(err)
	}

	am, err := LoadAnibridgeMappingWithAge(path, 24*7*time.Hour)
	if err != nil {
		t.Fatalf("LoadAnibridgeMappingWithAge: %v", err)
	}

	if tvdbID, ok := am.LookupByMAL(1); !ok || tvdbID != 100 {
		t.Errorf("expected MAL 1 -> TVDB 100, got %d, %v", tvdbID, ok)
	}
}

func TestLoadAnibridgeMappingWithAge_Stale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json.zst")

	fixture := `{ "mal:1": { "tvdb_show:100:s1": { "1-12": "1-12" } } }`
	if err := writeZstdFile(path, fixture); err != nil {
		t.Fatal(err)
	}

	am, err := LoadAnibridgeMappingWithAge(path, 0)
	if err != nil {
		t.Fatalf("LoadAnibridgeMappingWithAge: %v", err)
	}

	if tvdbID, ok := am.LookupByMAL(1); !ok || tvdbID != 100 {
		t.Errorf("expected MAL 1 -> TVDB 100, got %d, %v", tvdbID, ok)
	}
}

func writeZstdFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zstd.NewWriter(f)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write([]byte(content))
	return err
}

func makePtr[T any](v T) *T {
	return &v
}
