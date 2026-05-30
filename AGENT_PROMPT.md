# Implementation Agent Prompt: v2 Rewrite

**⚠️ READ THIS FILE, THEN DELETE IT. Do not leave this file on disk after reading.**

## Objective

Rewrite `anilistgen` from an MDBList list manager into a Sonarr-compatible TVDB list generator. No more MDBList API calls for list CRUD. Output is static JSON files for GitHub Pages.

## Architecture

Remove these packages entirely:
- `internal/mdblist/` — Replace with local mapping resolution
- `internal/season/` — Already removed
- `cmd/anilistgen/main.go` — rewrite

New structure:
```
cmd/anilistgen/main.go       — Simplified CLI: just fetch + resolve + output
internal/anilist/            — Keep as-is (GraphQL client)
internal/config/             — Keep but simplify (remove MDBList-specific fields)
internal/mapping/            — NEW: TVDB resolution from community/anime-lists files
internal/output/             — NEW: Sonarr JSON generation (gitHub Pages format)
internal/logging/            — Keep as-is
```

## Data Flow

```
main.go → FetchSeason → Filter (duration, blacklist, tags, ahead) 
        → Resolve (community mapping → anime-lists → optional MDBList fallback)
        → Output JSON → GitHub Pages
```

## Key Implementation Steps

### Step 1: New `internal/mapping/` package

Three files:

**community.go** — Load and query the shinkro/community-mapping YAML

```go
package mapping

// CommunityMapping holds MAL→TVDB mappings from tvdb-mal.yaml
type CommunityMapping struct {
    data map[int]int // MAL ID → TVDB ID
}

// LoadCommunityMapping downloads or reads the tvdb-mal.yaml file
func LoadCommunityMapping(path string) (*CommunityMapping, error)

// Lookup returns the TVDB ID for a given MAL ID
func (m *CommunityMapping) Lookup(malID int) (int, bool)
```

Source: `https://raw.githubusercontent.com/shinkro/community-mapping/main/tvdb-mal.yaml`
Format: YAML with `AnimeMap` array, each entry has `malid` and `tvdbid`
Auto-download if file doesn't exist locally. Cache at configurable path.

**animelists.go** — Parse and query the Anime-Lists/anime-lists XML

```go
// AnimeListsMapping holds AniDB→TVDB mappings from anime-list-full.xml
type AnimeListsMapping struct {
    data map[int]int // AniDB ID → TVDB ID
}

// LoadAnimeListsMapping downloads or reads the anime-list-full.xml file
func LoadAnimeListsMapping(path string) (*AnimeListsMapping, error)

// Lookup returns the TVDB ID for a given AniDB ID
func (m *AnimeListsMapping) Lookup(anidbID int) (int, bool)
```

Source: `https://raw.githubusercontent.com/Anime-Lists/anime-lists/master/anime-list-full.xml`
Format: XML with `<anime anidbid="..." tvdbid="..." defaulttvdbseason="...">` elements
~10,688 entries, 1.7MB

**resolve.go** — Resolution pipeline

```go
// Resolver chains multiple mapping sources
type Resolver struct {
    community *CommunityMapping  // fast path
    animelists *AnimeListsMapping // secondary path
}

// Resolve tries each mapping source in order
func (r *Resolver) Resolve(malID int, anidbID *int) (int, error)

// ResolveBatch resolves multiple MAL IDs in parallel
func (r *Resolver) ResolveBatch(shows []Show) map[int]int
```

Resolution order:
1. Community mapping (instant, local YAML) → 78% coverage
2. Anime-lists (via Jikan API for MAL→AniDB bridge) → ~15% more
3. MDBList batch lookup (optional, throttled) → remaining

### Step 2: New `internal/output/` package

```go
package output

// Show represents a resolved show for Sonarr output
type Show struct {
    TVDBID int    `json:"tvdbId"`
    Title  string `json:"title,omitempty"`
}

// SeasonOutput represents one season's list
type SeasonOutput struct {
    Season     string `json:"season"`
    Year       int    `json:"year"`
    GeneratedAt string `json:"generated_at"`
    Shows      []Show `json:"shows"`
}

// WriteSeasonJSON writes a Sonarr-compatible JSON file
func WriteSeasonJSON(dir, season string, year int, shows []Show) error

// WriteAllJSON writes all seasons as individual JSON files
func WriteAllJSON(outputDir string, seasonal map[string][]Show) error
```

Output format (per season file: `winter-2026.json`):
```json
{
  "season": "winter",
  "year": 2026,
  "generated_at": "2026-01-15T00:00:00Z",
  "shows": [
    {"tvdbId": 377543, "title": "JUJUTSU KAISEN Season 3"},
    {"tvdbId": 424536, "title": "Frieren: Beyond Journey's End Season 2"}
  ]
}
```

### Step 3: Simplify CLI (`cmd/anilistgen/main.go`)

```go
func main() {
    flag.String("output", "./sonarr-lists", "Output directory for JSON files")
    flag.String("config", "", "Config file path")
    flag.Bool("dry-run", false, "Print results without writing files")
    flag.Parse()
    
    // 1. Load config + community mapping + anime-lists
    // 2. Fetch all seasons from AniList
    // 3. Filter (duration, blacklist, tags, ahead_months)
    // 4. Resolve MAL IDs → TVDB IDs via mapping pipeline
    // 5. Write Sonarr JSON files per season
    // 6. Write manual_match.yml for unresolved shows
}
```

### Step 4: GitHub Actions Workflow (`.github/workflows/weekly-sync.yml`)

```yaml
name: Weekly anime list sync
on:
  schedule:
    - cron: '0 6 * * 0'
  workflow_dispatch:
  push:
    branches: [main]
    paths: ['.github/workflows/weekly-sync.yml', 'cmd/**', 'internal/**']

jobs:
  generate:
    runs-on: ubuntu-latest
    permissions:
      contents: write  # for gh-pages branch push
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go build ./cmd/anilistgen
      - run: ./anilistgen -output ./out
      - uses: peaceiris/actions-gh-pages@v3
        with:
          publish_dir: ./out
          publish_branch: gh-pages
```

### Step 5: Remove Unused Dependencies

- Remove `gopkg.in/yaml.v3` dependency (community mapping uses it, keep it)
- Remove `internal/mdblist/` package (entire directory)
- Simplify `internal/config/` — remove MDBList-specific fields:
  - Remove: `MDBListAPIKey`, `MDBList.TitleTemplate`, `MDBList.DescriptionTemplate`, `MDBList.Public`, `MDBList.Blacklist`
  - Keep: `AniList` settings, `Interval`, `Logging`, `StateFile`

### Step 6: Update Config Structure

```yaml
# Simplified config
anilist:
  years: [2026]
  seasons: [all]
  max_per_season: 100
  include_ona: false      # default false (TV only, no ONA)
  winter_overflow: true
  ahead_months: 3
  exclude_tags:
    - "Hentai"

# Filter
blacklist: []

# Output
output_dir: ./sonarr-lists

# Mapping file paths (auto-downloaded if missing)
community_mapping_path: /tmp/anilistgen_tvdb.yaml
anime_lists_path: /tmp/anime-list-full.xml

# Optional Sonarr API push
sonarr:
  url: ""
  api_key: ""
  quality_profile: "HD-1080p"
  root_folder: "/tv"
```

### Step 7: Testing

Create a simple integration test:
```bash
go build ./cmd/anilistgen
./anilistgen -dry-run -output /tmp/test-output
# Verify: ls /tmp/test-output/ shows JSON files
# Verify: jq '.shows | length' /tmp/test-output/winter-2026.json
```

## Key Estimates

| Metric | v1 (MDBList) | v2 (direct) |
|---|---|---|
| Sync time | ~5 minutes | ~10-15 seconds |
| API calls | ~60 per sync | 4-6 (AniList only) |
| Rate limits | MDBList 10K/day | None (static files) |
| Dependencies | MDBList API | Community YAML + XML |
| Output format | MDBList lists | JSON → GitHub Pages → Sonarr |

## Files to Create

1. `internal/mapping/community.go`
2. `internal/mapping/animelists.go`
3. `internal/mapping/resolve.go`
4. `internal/output/output.go`
5. `cmd/anilistgen/main.go` (rewrite)
6. `internal/config/config.go` (simplify)
7. `.github/workflows/weekly-sync.yml`
8. `internal/filter/filter.go` (extract from old sync)

## Files to Remove

1. `internal/mdblist/` (entire package)
2. `internal/sync/sync.go` (replace with mapping + output)
3. Old `Dockerfile`, `Dockerfile.goreleaser`, `.goreleaser.yaml` (simplify for GitHub Actions)

## Implementation Order

1. Create `internal/mapping/` with community mapping loader
2. Create `internal/output/` with Sonarr JSON writer
3. Simplify `internal/config/` (drop MDBList fields)
4. Rewrite `cmd/anilistgen/main.go` (new pipeline)
5. Remove `internal/mdblist/` and `internal/sync/`
6. Add GitHub Actions workflow
7. Test with dry-run against live AniList
8. Push and verify GH Pages output
