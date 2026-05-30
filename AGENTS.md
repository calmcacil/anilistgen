# anilistgen — Agent Guide

## Overview

A Go CLI tool that fetches seasonal anime from **AniList** (GraphQL API) and
generates **Sonarr-compatible JSON** files with TVDB IDs, hosted on GitHub
Pages for direct Sonarr import.

## Architecture

```
cmd/anilistgen/main.go    — Simplified CLI: fetch + resolve + output
internal/anilist/            — AniList GraphQL client (retry/backoff, pagination)
internal/config/             — YAML config loading, validation, init-config gen
internal/mapping/            — TVDB resolution from anime-lists + community mapping
internal/filter/             — Show filtering (duration, blacklist, tags, ahead)
internal/output/             — Sonarr compact JSON generation
internal/logging/            — slog setup
```

## Subcommands

| Command | Description |
|---|---|
| `anilistgen` | Generate JSON files (default) |
| `anilistgen init-config` | Generate default YAML config |
| `anilistgen validate` | Check config + AniList connectivity |

## External APIs

### AniList GraphQL

- **Endpoint**: `POST https://graphql.anilist.co` — no auth
- **Query**: Fetches `Media` with filters: `type: ANIME`, `season`, `seasonYear`, `format_in: [TV, ONA]`
- **Pagination**: Paginates through AniList's 50-per-page limit up to `max_per_season`
- **Retry**: 3 attempts with exponential backoff (1s, 2s, 4s) on non-200 responses
- **Fields returned**: `id`, `idMal`, `title.{romaji,english}`, `format`, `episodes`, `duration`, `genres`, `status`, `startDate`

### No MDBList API

The v2 rewrite removes all MDBList dependencies. TVDB resolution uses local
mapping files instead.

## Data Sources (Mapping)

### Anime-Lists (primary)

- **Source**: Anime-Lists/anime-lists (`anime-list-full.xml`)
- **Mapping**: AniList ID → TVDB ID (1.7MB, ~10,688 entries)
- **Auto-downloaded** on first run, cached locally

### Community mapping (secondary)

- Source: shinkro/community-mapping (`tvdb-mal.yaml`)
- **Mapping**: MAL ID → TVDB ID (947KB, ~5,241 entries)
- Covers ~78% of seasonal anime as fallback

## Sync Pipeline

For each season/year:

1. **Fetch** — Query AniList for TV/ONA anime (paginated)
2. **Winter overflow** — For WINTER season, fetches prior year's WINTER and
   merges only shows with `startDate.month == 12` (December premieres tagged
   under the prior calendar year)
3. **Filter** — Remove shows with duration ≤10 min, blacklisted shows,
   tag-excluded shows, and shows too far in the future
4. **Resolve** — Try anime-lists (AniList ID → TVDB) first, then community
   mapping (MAL ID → TVDB) as fallback
5. **Output** — Write compact JSON per season + yearly aggregate

## Config File

Location (searched in order):
1. `-config` CLI flag path
2. `./anilistgen.yaml`
3. `$XDG_CONFIG_HOME/anilistgen/anilistgen.yaml` (defaults to `~/.config/...`)

Unknown top-level keys produce a warning on stderr but don't prevent startup.

## Output Format

Per season: `winter-2026.json`
Yearly: `2026.json`

```json
{"season":"winter","year":2026,"generated_at":"...","shows":[{"tvdbId":377543,"title":"..."}]}
```

JSON is minified (no whitespace) for faster Sonarr import.

## GitHub Actions

Workflow at `.github/workflows/weekly-sync.yml`:
- Runs weekly (Sunday 6 AM) + on push to main
- Builds binary, generates JSON, publishes to `gh-pages` branch
- Result served at `https://{user}.github.io/anilistgen/{season}-{year}.json`

## Testing

```bash
go build ./cmd/anilistgen
go vet ./...
go test ./...
./anilistgen -dry-run
./anilistgen -output /tmp/x
./anilistgen validate
```

## Security Notes

- **Config in .gitignore** — prevents accidental commits
- **No auth on AniList reads** — AniList GraphQL is public, no credentials needed
- **No API keys in output** — all mapping is local file-based
