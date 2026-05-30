# Sonarr Seasonal Anime Lists

Generate weekly Sonarr-compatible anime season lists from AniList data, hosted on GitHub Pages for direct Sonarr import.

## Goal

A zero-infrastructure GitHub Action that:
1. Fetches currently-airing seasonal anime from AniList
2. Resolves each show to its TVDB ID
3. Outputs a Sonarr-compatible JSON file per season
4. Hosts the files on GitHub Pages for Sonarr to import

## Data Sources

### Primary: community-mapping (shinkro)
- **Source**: https://github.com/shinkro/community-mapping
- **File**: `tvdb-mal.yaml` (947KB, 5,241 entries)
- **Coverage**: 78% of seasonal anime (155/198 shows tested)
- **Format**: Direct MAL ID → TVDB ID mapping
- **Access**: Download raw YAML at build time

### Secondary: Anime-Lists (comprehensive)
- **Source**: https://github.com/Anime-Lists/anime-lists
- **File**: `anime-list-full.xml` (1.7MB, 10,688 entries)
- **Format**: AniDB ID → TVDB ID (needs MAL→AniDB conversion via Jikan API)
- **Access**: Parse XML locally, use Jikan for MAL→AniDB bridge
- **Usage**: Fallback for shows not in community-mapping

### Resolution Pipeline

```
AniList Show
  ├─ mal_id (from AniList API: idMal)
  │
  ├── Check: community-mapping (tvdb-mal.yaml)
  │     → TVDB ID found? → Done (instant, 78%)
  │
  ├── Fallback: Anime-Lists → Jikan API for MAL→AniDB → XML lookup
  │     → TVDB ID found? → Done (~1 API call, ~15%)
  │
  └── Fallback: MDBList API batch lookup
        → TVDB ID found? → Done (~7% remaining)
  
  All matched → Write to Sonarr JSON
  Not matched → Write to manual_match.yml
```

### Mapping Coverage Research

| Source | Entries | Key | TVDB? | TMDB? | Access |
|---|---|---|---|---|---|
| community-mapping | 5,241 | MAL→TVDB | ✅ Yes | ❌ No | YAML file |
| anime-lists | 10,688 | AniDB→TVDB | ✅ Yes | ✅ Yes | XML file |
| anime-offline-database | 40,921 | Multi→Multi | ❌ Rare | ❌ Rare | JSON file |
| MDBList API | ~100K+ | Multi→Multi | ✅ Yes | ✅ Yes | REST API (throttled) |

**Recommendation:** community-mapping + anime-lists covers >95% of seasonal anime. MDBList API only needed as last resort.

### GitHub Pages Endpoints

```
https://{user}.github.io/anilistgen/winter-2026.json
https://{user}.github.io/anilistgen/spring-2026.json
https://{user}.github.io/anilistgen/summer-2026.json
https://{user}.github.io/anilistgen/fall-2026.json
```

### Sonarr Import Format

Each JSON file is a Sonarr-compatible custom list:

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

Sonarr's "Custom List" import accepts an array of `{"tvdbId": 12345}` objects. The title field is for human readability only.

## Filters

- **Duration**: Skip shows with per-episode duration ≤ 10 minutes (shorts)
- **Format**: TV only (exclude ONA, OVA, Movie, Special)
- **Blacklist**: MAL ID or title substring (configurable)
- **Ahead months**: Skip shows starting more than N months ahead (default: 3)

## Output

### Sonarr JSON files (per season)
Written to the repository's output directory, then published to GitHub Pages.

### manual_match.yml
Shows that couldn't be resolved to any TVDB ID. User fills in the TVDB ID manually.

## GitHub Actions Workflow

```yaml
name: Weekly anime list sync
on:
  schedule:
    - cron: '0 6 * * 0'   # Every Sunday at 6 AM
  workflow_dispatch:       # Manual trigger

jobs:
  generate:
    runs-on: ubuntu-latest
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

## Configuration

```yaml
# How many months ahead to include shows (prevents unresolved future shows)
ahead_months: 3

# AniList query settings
anilist:
  years: [2026]
  seasons: [all]

# Filter
blacklist: []
exclude_tags: []

# Output
output_dir: ./sonarr-lists

# Optional: Direct Sonarr API push (instead of GitHub Pages)
sonarr:
  url: ""
  api_key: ""
  quality_profile: "HD-1080p"
  root_folder: "/tv"
```

## Architecture

```
cmd/anilistgen/main.go      — Entry point (simplified CLI)
internal/anilist/           — AniList GraphQL client (unchanged)
internal/config/            — YAML config (simplified)
internal/mapping/           — TVDB mapping resolution (NEW)
  ├── community.go          — Load tvdb-mal.yaml from shinkro
  ├── animelists.go         — Parse anime-list-full.xml from Anime-Lists
  └── resolve.go            — Resolution pipeline orchestrator
internal/filter/            — Show filtering (duration, blacklist, tags, ahead)
internal/output/            — Sonarr JSON generation (NEW)
internal/logging/           — slog setup (unchanged)
```

## Key Decisions

- **No MDBList dependency**: The mapping files replace MDBList API for TVDB resolution
- **Static hosting**: GitHub Pages is free, zero-infrastructure, always available
- **Weekly sync**: AniList's seasonal data doesn't change daily; weekly is sufficient
- **No server**: Everything is static files + GitHub Actions
- **Sonarr-native format**: Sonarr's Custom List import supports JSON with tvdbId array
