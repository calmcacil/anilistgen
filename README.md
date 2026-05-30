# animelistgen

> Fetch all anime for every season of a given year from AniList and
> create/update corresponding MDBList lists for use with Agregarr.

## Why this exists

Agregarr has no built-in "all shows airing this season" source. Community
Trakt/MDBList lists for seasonal anime are unreliable. This tool generates
authoritative MDBList lists directly from AniList's seasonal data so you can
paste the URLs into Agregarr as MDBList Custom List sources.

## Spec

See [`specs/anilist-seasonal-mdblist/PRODUCT.md`](specs/anilist-seasonal-mdblist/PRODUCT.md)
for the full behavioral specification.

## Usage

```bash
# Generate a default config
animelistgen init-config

# One-shot: process all configured seasons
animelistgen

# Validate config and API connectivity
animelistgen validate

# Background daemon with configurable interval
animelistgen daemon

# Dry-run (print what would be done without API calls)
animelistgen -dry-run
```

## Configuration

The tool reads from a YAML config file at `~/.config/animelistgen/animelistgen.yaml`
(or `./animelistgen.yaml`, or a path specified via `-config`).

See `animelistgen init-config` to generate a default config with all options
documented inline.

## Deployment options

### One-shot (cron / systemd timer)

Install the binary, write a config, and run via systemd:

```bash
# Install
sudo cp animelistgen /usr/local/bin/
sudo mkdir -p /etc/animelistgen
sudo cp animelistgen.yaml /etc/animelistgen/

# Enable daily timer
sudo cp deploy/animelistgen.service /etc/systemd/system/
sudo cp deploy/animelistgen.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now animelistgen.timer
```

### Daemon (Docker)

```bash
docker compose up -d
```

### Agregarr integration

After each sync, paste the printed MDBList URLs into Agregarr as **MDBList →
Custom List** sources. Create one collection per season. Re-runs update the
lists in-place — Agregarr picks up changes on its next sync.

## Architecture

```
cmd/animelistgen/main.go    — entry point, subcommand dispatch
internal/config/            — config struct, loading, YAML parsing
internal/anilist/           — AniList GraphQL client
internal/mdblist/           — MDBList API client (create/update lists)
internal/season/            — season detection and constants
internal/logging/           — structured logging setup
internal/sync/              — sync orchestration (fetch + publish)
deploy/                     — systemd units, Dockerfile, docker-compose
```

## Dependencies

- AniList GraphQL API — no auth needed for reads.
- MDBList API key — required for writes (get one at https://mdblist.com/api).

---

## Session context for a new AI agent

### What needs to be built

A Go CLI tool + daemon that:

1. Reads a YAML config file defining which seasons to fetch, MDBList
   credentials, list naming templates, and sync interval.
2. Queries AniList's GraphQL API for all TV/ONA anime per season/year.
3. Creates or updates MDBList lists with those shows (diff-based).
4. Supports one-shot, daemon, init-config, and validate subcommands.
5. Ships sample systemd unit/timer and Dockerfile.

The full behavioral spec is in `specs/anilist-seasonal-mdblist/PRODUCT.md`
— read it first before implementing.

### Current state

- Repo is scaffolded with Go module and stub packages.
- `internal/anilist/` — GraphQL client with retry/backoff.
- `internal/mdblist/` — MDBList client (list, find, create, update).
- `internal/season/` — season detection helpers.
- `internal/config/` — currently has flag-based config from the old design;
  needs to be replaced with YAML-file-based config per the new spec.
- `cmd/animelistgen/main.go` — entry point stubbed out; needs subcommand
  dispatch and orchestration.

### External APIs

- **AniList GraphQL**: `POST https://graphql.anilist.co` — no auth.
  Variables: `MediaSeason`, `Int`, pagination.
- **MDBList API**: requires API key. Endpoints:
  - `GET /api/lists` — list user's lists
  - `POST /api/list` — create a list
  - `PUT /api/list/{id}` — update list items
  - See https://mdblist.com/api

### Implementation order (recommended)

1. **Config** — replace `internal/config/` with YAML-based loading using
   `gopkg.in/yaml.v3`. Write `init-config` subcommand.
2. **Sync** — new `internal/sync/` package that orchestrates AniList fetch →
   MDBList publish per season, with diff comparison.
3. **Subcommands** — update `cmd/animelistgen/main.go` with subcommand
   dispatch (oneshot, daemon, init-config, validate).
4. **Daemon** — loop with configurable interval, signal handling.
5. **Deploy** — `deploy/` directory with systemd unit/timer, Dockerfile,
   docker-compose.yml.
6. **Validate** — connectivity checks for both APIs.
