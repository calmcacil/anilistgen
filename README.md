# animelistgen

Fetch all anime for every season of a given year from AniList and
create/update corresponding MDBList lists for use with Agregarr.

## Why

Agregarr has no built-in "all shows airing this season" source. Community
Trakt/MDBList lists for seasonal anime are unreliable. This tool generates
authoritative MDBList lists directly from AniList's seasonal data.

## Quick start

```bash
# 1. Generate a default config
./animelistgen init-config

# 2. Edit config — add your MDBList API key and years
#    (get a free key at https://mdblist.com)
vim animelistgen.yaml

# 3. Dry-run to preview what will happen
./animelistgen -dry-run

# 4. Push lists to MDBList
./animelistgen
```

## Usage

```bash
animelistgen              Oneshot — process all seasons, print URLs
animelistgen daemon       Daemon — loop at configurable interval
animelistgen init-config  Generate a default config file
animelistgen validate     Validate config and API connectivity
```

### Global flags

| Flag | Description |
|---|---|
| `-config PATH` / `-c PATH` | Config file path (overrides default search) |
| `-dry-run` | Print what would be done without MDBList API calls |
| `-output DIR` / `-o DIR` | Write JSON files per season instead of MDBList |
| `-v` / `-verbose` | Verbose logging |
| `-h` / `-help` | Print help |

## Configuration

The tool reads config from these sources (later overrides earlier):
1. Config file: `./animelistgen.yaml` → `~/.config/animelistgen/animelistgen.yaml`
2. CLI flag: `-config PATH`
3. **Environment variables** — every setting has an `ALG_` prefixed env var

Run `animelistgen init-config` to generate a default file with all options
documented inline. Key settings:

```yaml
mdblist_api_key: ""          # Required. Get one at https://mdblist.com
anilist:
  years: [2026, 2027]        # Years to process
  seasons: [all]             # Or: winter, spring, summer, fall
  max_per_season: 100
  include_ona: true          # Include ONA format alongside TV
mdblist:
  title_template: "Anime {season} {year}"
  public: true
  blacklist: []              # Titles or MAL IDs to skip
```

### Environment variables

Every config field can be set via `ALG_` prefixed env vars — no config file
needed. This is especially useful for Docker.

| Env var | Maps to | Example |
|---|---|---|
| `ALG_MDBLIST_API_KEY` | `mdblist_api_key` | `abc123` |
| `ALG_INTERVAL` | `interval` | `24h` |
| `ALG_ANILIST_YEARS` | `anilist.years` | `2026,2027` |
| `ALG_ANILIST_SEASONS` | `anilist.seasons` | `winter,spring` |
| `ALG_ANILIST_MAX_PER_SEASON` | `anilist.max_per_season` | `100` |
| `ALG_ANILIST_INCLUDE_ONA` | `anilist.include_ona` | `true` |
| `ALG_MDBLIST_TITLE_TEMPLATE` | `mdblist.title_template` | `Anime {season} {year}` |
| `ALG_MDBLIST_DESCRIPTION_TEMPLATE` | `mdblist.description_template` | `...` |
| `ALG_MDBLIST_PUBLIC` | `mdblist.public` | `true` |
| `ALG_MDBLIST_BLACKLIST` | `mdblist.blacklist` | `One Piece,57658` |
| `ALG_LOG_LEVEL` | `logging.level` | `info` |

Legacy `MDBLIST_API_KEY` is still supported as a fallback for the API key.

## MDBList plan limits

The free MDBList plan caps you at **4 static lists**. If you sync
4 seasons × 2 years = 8 lists, you'll hit this limit on the 5th list.

| Plan | Lists | Upgrade |
|---|---|---|
| Free | 4 | — |
| 1€/month | More | [Patreon](https://www.patreon.com/mdblist) |

## Agregarr integration

After a sync, paste the printed MDBList URLs into Agregarr as **MDBList →
Custom List** sources. Create one collection per season. Re-running updates
the lists in-place — Agregarr picks up changes on its next sync.

## Deployment

### One-shot via systemd timer

```bash
sudo cp animelistgen /usr/local/bin/
sudo cp deploy/animelistgen.service /etc/systemd/system/
sudo cp deploy/animelistgen.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now animelistgen.timer
```

### Docker (ghcr.io)

A prebuilt image is available at `ghcr.io/calmcacil/animelistgen`.
Every push to `master` and every semver tag rebuilds it automatically.

**Daemon mode** (default):

```bash
docker run -d --name animelistgen --restart unless-stopped \
  -e ALG_MDBLIST_API_KEY=your_key_here \
  -e ALG_INTERVAL=24h \
  -e ALG_ANILIST_YEARS=2026,2027 \
  ghcr.io/calmcacil/animelistgen:latest
```

**One-shot mode**:

```bash
docker run --rm \
  -e ALG_MDBLIST_API_KEY=your_key_here \
  -e ALG_INTERVAL=0 \
  ghcr.io/calmcacil/animelistgen:latest animelistgen
```

**Using docker-compose** (edit `docker-compose.yml` or set env in `.env`):

```bash
docker compose up -d
```

All configuration is via environment variables — no config file mount required.
If you prefer, you can mount a YAML config at `/etc/animelistgen/animelistgen.yaml`
and omit the env vars.

## Output modes

| Mode | Trigger | What happens |
|---|---|---|
| Normal | `animelistgen` | Creates/updates MDBList lists, prints URLs |
| Dry-run | `animelistgen -dry-run` | Fetches AniList, prints what it would do, no MDBList calls |
| File | `animelistgen -output /tmp/x` | Writes JSON files per season instead of MDBList |

## Contributing

1. Read [`AGENTS.md`](./AGENTS.md) for architecture details.
2. Read [`specs/anilist-seasonal-mdblist/PRODUCT.md`](./specs/anilist-seasonal-mdblist/PRODUCT.md)
   for the full behavioral specification.
3. Run `go vet ./...` before submitting changes.
4. The config file `animelistgen.yaml` is in `.gitignore` — use
   `animelistgen.yaml.example` as a reference. Never commit real API keys.
5. `go build ./cmd/animelistgen` produces the binary.
