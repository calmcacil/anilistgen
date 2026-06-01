package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/calmcacil/anilistgen/internal/anilist"
	"github.com/calmcacil/anilistgen/internal/mapping"
)

func TestResolveBatch(t *testing.T) {
	// Create a test CommunityMapping via temp YAML file
	dir := t.TempDir()
	path := filepath.Join(dir, "tvdb-mal.yaml")
	content := `AnimeMap:
  - malid: 16498
    tvdbid: 12345
  - malid: 99999
    tvdbid: 67890
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cm, err := mapping.LoadCommunityMapping(path)
	if err != nil {
		t.Fatalf("LoadCommunityMapping: %v", err)
	}
	resolver := mapping.NewResolver(cm)

	// Test with empty map - dry run
	result := resolveBatch(resolver, map[string][]anilist.Show{}, true)
	if len(result) != 0 {
		t.Errorf("expected empty result for dry-run, got %d entries", len(result))
	}

	// Test with empty map - non-dry run
	result = resolveBatch(resolver, map[string][]anilist.Show{}, false)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d entries", len(result))
	}

	// Test with invalid key format
	result = resolveBatch(resolver, map[string][]anilist.Show{
		"invalid-key": nil,
	}, false)
	if len(result) != 0 {
		t.Errorf("expected empty result for invalid key, got %d entries", len(result))
	}

	// Test with valid key but nil shows (should produce empty output slice)
	result = resolveBatch(resolver, map[string][]anilist.Show{
		"WINTER-2026": nil,
	}, false)
	if shows, ok := result["WINTER-2026"]; !ok {
		t.Error("expected WINTER-2026 key in result")
	} else if len(shows) != 0 {
		t.Errorf("expected 0 shows for nil input, got %d", len(shows))
	}

	// Test resolution with unresolvable show (IDMal not in mapping)
	result = resolveBatch(resolver, map[string][]anilist.Show{
		"WINTER-2026": {{ID: 1, IDMal: nil}},
	}, false)
	if shows, ok := result["WINTER-2026"]; ok && len(shows) != 0 {
		t.Errorf("expected 0 resolved shows for no IDMal, got %d", len(shows))
	}

	// Test resolution with resolvable show
	result = resolveBatch(resolver, map[string][]anilist.Show{
		"WINTER-2026": {{ID: 1, IDMal: makePtr(16498)}},
	}, false)
	if shows, ok := result["WINTER-2026"]; !ok {
		t.Error("expected WINTER-2026 key")
	} else if len(shows) != 1 {
		t.Errorf("expected 1 resolved show, got %d", len(shows))
	} else if shows[0].TVDBID != 12345 {
		t.Errorf("expected TVDB 12345, got %d", shows[0].TVDBID)
	}

	// Test dry-run output (captures stdout, just verify no panic and correct output)
	t.Run("dry-run output format", func(t *testing.T) {
		shows := []anilist.Show{
			{ID: 1, IDMal: makePtr(16498), Title: anilist.Title{English: makePtr("Test Show")}},
			{ID: 2, IDMal: nil},
		}
		result := resolveBatch(resolver, map[string][]anilist.Show{
			"WINTER-2026": shows,
		}, true)
		if len(result) != 0 {
			t.Error("expected empty result for dry run output")
		}
	})
}

func TestGroupBySeason(t *testing.T) {
	t.Parallel()

	winter := anilist.Show{ID: 1, Season: makePtr("WINTER")}
	spring := anilist.Show{ID: 2, Season: makePtr("SPRING")}
	summer := anilist.Show{ID: 3, Season: makePtr("SUMMER")}
	fall := anilist.Show{ID: 4, Season: makePtr("FALL")}
	unknown := anilist.Show{ID: 5, Season: nil}
	lower := anilist.Show{ID: 6, Season: makePtr("winter")}

	result := groupBySeason([]anilist.Show{winter, spring, summer, fall, unknown, lower})

	if len(result["WINTER"]) != 2 {
		t.Errorf("expected 2 WINTER shows, got %d", len(result["WINTER"]))
	}
	if len(result["SPRING"]) != 1 {
		t.Errorf("expected 1 SPRING show, got %d", len(result["SPRING"]))
	}
	if len(result["SUMMER"]) != 1 {
		t.Errorf("expected 1 SUMMER show, got %d", len(result["SUMMER"]))
	}
	if len(result["FALL"]) != 1 {
		t.Errorf("expected 1 FALL show, got %d", len(result["FALL"]))
	}
	if len(result["UNKNOWN"]) != 1 {
		t.Errorf("expected 1 UNKNOWN show, got %d", len(result["UNKNOWN"]))
	}

	// ID 6 (lowercase "winter") should be in WINTER via SeasonCode()
	found := false
	for _, s := range result["WINTER"] {
		if s.ID == 6 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected lowercase winter show in WINTER bucket")
	}
}

func TestFilterDecember(t *testing.T) {
	t.Parallel()

	dec := anilist.Show{ID: 1, StartDate: anilist.FuzzyDate{Month: makePtr(12)}}
	jan := anilist.Show{ID: 2, StartDate: anilist.FuzzyDate{Month: makePtr(1)}}
	nilMonth := anilist.Show{ID: 3, StartDate: anilist.FuzzyDate{Month: nil}}

	all := []anilist.Show{{ID: 10}}
	added := filterDecember(&all, []anilist.Show{dec, jan, nilMonth})

	if added != 1 {
		t.Errorf("expected 1 added (December), got %d", added)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 shows total, got %d", len(all))
	}
	if all[1].ID != 1 {
		t.Errorf("expected added show to have ID 1, got %d", all[1].ID)
	}
}

func TestFilterDecember_Deduplicates(t *testing.T) {
	t.Parallel()

	dec := anilist.Show{ID: 1, StartDate: anilist.FuzzyDate{Month: makePtr(12)}}
	all := []anilist.Show{dec}
	added := filterDecember(&all, []anilist.Show{dec})

	if added != 0 {
		t.Errorf("expected 0 added (already present), got %d", added)
	}
}

func makePtr[T any](v T) *T {
	return &v
}
