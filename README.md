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

The tool reads `animelistgen.yaml` from `./`, then
`~/.config/animelistgen/animelistgen.yaml`, or any path passed via `-config`.

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

The API key can also be set via the `MDBLIST_API_KEY` environment variable,
which takes precedence when the config value is empty.

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

### Daemon via Docker

```bash
docker compose up -d
```

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
