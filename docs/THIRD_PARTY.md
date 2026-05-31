# Third-Party Licenses

anilistgen uses the following third-party data sources, libraries,
and tools. Each is listed with its license and how it is used.

## Data Sources

### AniList API
- **License:** Terms of Use (non-commercial free, commercial < $150/mo free)
- **Usage:** Primary data source — queried at runtime via GraphQL API
- **Website:** https://anilist.co
- **Terms:** https://docs.anilist.co/guide/terms-of-use

### shinkro/community-mapping
- **License:** MIT License
- **Copyright:** Copyright (c) 2023 Rohit Vardam
- **Usage:** TVDB ID resolution — auto-downloaded YAML mapping file
- **Repository:** https://github.com/shinkro/community-mapping
- **Attribution:** See [NOTICE](../NOTICE) at the project root.

## Go Dependencies

### gopkg.in/yaml.v3
- **License:** MIT License
- **Copyright:** Copyright (c) 2011-2019 Canonical Ltd
- **Usage:** YAML config and mapping file parsing
- **Source:** https://github.com/go-yaml/yaml

## CI/CD Tools

### actions/checkout
- **License:** MIT License
- **Usage:** GitHub Actions — checkout repository
- **Repository:** https://github.com/actions/checkout

### actions/setup-go
- **License:** MIT License
- **Usage:** GitHub Actions — install Go toolchain
- **Repository:** https://github.com/actions/setup-go

### peaceiris/actions-gh-pages
- **License:** MIT License
- **Usage:** GitHub Actions — deploy to GitHub Pages
- **Repository:** https://github.com/peaceiris/actions-gh-pages
