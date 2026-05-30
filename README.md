# anilistgen

Generate Sonarr-compatible seasonal anime lists from **AniList** data, hosted
on GitHub Pages for direct Sonarr import.

Sonarr has no built-in "all shows airing this season" source. This tool
generates authoritative JSON lists directly from AniList's seasonal data,
resolves each show to its TVDB ID, and publishes static files for Sonarr
to consume as a Custom List.

---

## Quick start

```bash
# Generate config
./anilistgen init-config

# Preview what will be generated
./anilistgen -dry-run

# Generate JSON files
./anilistgen -output ./sonarr-lists

# Validate config and API connectivity
./anilistgen validate
```

---

## Reference

### Commands

| Command | Description |
|---|---|
| `anilistgen` | Generate JSON files (default) |
| `anilistgen init-config` | Generate a default YAML config file |
| `anilistgen validate` | Validate config + test AniList connectivity |

### Flags

| Flag | Description |
|---|---|
| `-config PATH`, `-c PATH` | Config file path (overrides default search) |
| `-dry-run` | Print results without writing files |
| `-output DIR`, `-o DIR` | Output directory (overrides config) |
| `-v`, `-verbose` | Verbose (debug) logging |
| `-h`, `-help` | Print usage |
| `-version`, `-V` | Print version |

---

## Configuration

The tool reads settings from three sources (later overrides earlier):

1. **Config file**: `./anilistgen.yaml` → `~/.config/anilistgen/anilistgen.yaml`
2. **CLI flag**: `-config /path/to/config.yaml`
3. **Environment variables**: Every setting has an `ALG_` prefixed env var

### Config file reference

Run `anilistgen init-config` to generate a documented template.

```yaml
anilist:
  years: [2026]
  seasons: [all]
  max_per_season: 100
  include_ona: false
  winter_overflow: true
  ahead_months: 3
  exclude_tags: []

blacklist: []

output_dir: ./sonarr-lists

community_mapping_path: /tmp/anilistgen_tvdb.yaml
anime_lists_path: /tmp/anime-list-full.xml

logging:
  level: info
  file: ""
```

### Environment variables

| Env var | Maps to | Default |
|---|---|---|
| `ALG_ANILIST_YEARS` | `anilist.years` | current year |
| `ALG_ANILIST_SEASONS` | `anilist.seasons` | `all` |
| `ALG_ANILIST_MAX_PER_SEASON` | `anilist.max_per_season` | `100` |
| `ALG_ANILIST_INCLUDE_ONA` | `anilist.include_ona` | `false` |
| `ALG_ANILIST_WINTER_OVERFLOW` | `anilist.winter_overflow` | `true` |
| `ALG_ANILIST_EXCLUDE_TAGS` | `anilist.exclude_tags` | `""` |
| `ALG_ANILIST_AHEAD_MONTHS` | `anilist.ahead_months` | `3` |
| `ALG_BLACKLIST` | `blacklist` | `""` |
| `ALG_OUTPUT_DIR` | `output_dir` | `./sonarr-lists` |
| `ALG_COMMUNITY_MAPPING_PATH` | `community_mapping_path` | `/tmp/anilistgen_tvdb.yaml` |
| `ALG_ANIME_LISTS_PATH` | `anime_lists_path` | `/tmp/anime-list-full.xml` |
| `ALG_LOG_LEVEL` | `logging.level` | `info` |
| `ALG_LOG_FILE` | `logging.file` | `""` (stderr) |

**Notes on env var format:**
- Lists use comma separation: `2026,2027`
- Booleans accept `true`/`1` or `false`/`0`

### Filters

- **Duration** — Skips shows with per-episode duration ≤ 10 minutes
- **Format** — TV only by default; ONA can be included via `include_ona: true`
- **Blacklist** — MAL ID or title substring (case-insensitive)
- **Tag exclusion** — AniList content tags to skip (e.g. `"Hentai"`)
- **Ahead months** — Skips shows starting more than N months in the future

---

## Output format

### Per-season files: `winter-2026.json`

```json
{"season":"winter","year":2026,"generated_at":"2026-01-15T00:00:00Z","shows":[{"tvdbId":377543,"title":"JUJUTSU KAISEN Season 3"},{"tvdbId":424536,"title":"Frieren: Beyond Journey's End Season 2"}]}
```

### Yearly file: `2026.json`

Same format without the `season` field. Aggregates all resolved shows for the year.

Sonarr's Custom List import accepts `{"tvdbId": 12345}` entries. The `title`
field is for human readability only. JSON is minified for faster imports.

---

## Data sources

### Primary: Anime-Lists (anime-lists)

- **Source**: [Anime-Lists/anime-lists](https://github.com/Anime-Lists/anime-lists)
- **File**: `anime-list-full.xml` — 1.7MB, ~10,688 entries
- **Mapping**: AniList ID → TVDB ID
- **Access**: Downloaded on first run, cached locally

### Secondary: community-mapping (shinkro)

- **Source**: [shinkro/community-mapping](https://github.com/shinkro/community-mapping)
- **File**: `tvdb-mal.yaml` — 947KB, ~5,241 entries
- **Mapping**: MAL ID → TVDB ID
- **Coverage**: ~78% of seasonal anime
- **Access**: Downloaded on first run, cached locally

### Resolution pipeline

```
AniList Show
  └─ anilist_id (from API)
       │
       ├── Anime-Lists (instant, local XML)  → TVDB ID? → Done (~75%)
       │
       └── Community mapping (instant, local YAML) → TVDB ID? → Done (~78% of remainder)
       
  Not matched → silently skipped (not yet in TVDB)
```

---

## How it works

### Pipeline

```
AniList query → Winter overflow (merge Dec shows from prior year)
  → Filter (duration, blacklist, tags, ahead months)
  → Resolve (anime-lists → community mapping)
  → Write JSON files (one per season + yearly aggregate)
```

### Winter overflow

AniList tags WINTER shows by calendar year. A December 2025 premiere is
tagged as WINTER 2025, not WINTER 2026. With `winter_overflow: true`, the
tool fetches the prior year's WINTER and merges December-premiering shows
into the current WINTER list.

---

## GitHub Actions

A workflow (`.github/workflows/weekly-sync.yml`) runs every Sunday at 6 AM,
builds the tool, generates JSON files, and publishes them to GitHub Pages.

```yaml
on:
  schedule:
    - cron: '0 6 * * 0'
  workflow_dispatch:
```

Output is published to the `gh-pages` branch, available at:
```
https://{user}.github.io/anilistgen/winter-2026.json
https://{user}.github.io/anilistgen/2026.json
```

---

## Sonarr integration

Add a Custom List in Sonarr pointing to the GitHub Pages URL:
```
Settings → Import Lists → Add → Custom List
  URL: https://{user}.github.io/anilistgen/winter-2026.json
```

Sonarr refreshes lists periodically and imports any shows it doesn't
already have.

---

## Season timing

| Season   | Months | Typical premiere window |
|---|---|---|
| **WINTER** | Dec–Feb | Early January |
| **SPRING** | Mar–May | Early April |
| **SUMMER** | Jun–Aug | Early July |
| **FALL** | Sep–Nov | Early October |

---

## Contributing

1. Read [`AGENTS.md`](./AGENTS.md) for architecture and design.
2. Run `go vet ./...` and `go test ./...` before submitting changes.
3. Config files are in `.gitignore` — never commit real API keys.
4. Build with `go build ./cmd/anilistgen`.
