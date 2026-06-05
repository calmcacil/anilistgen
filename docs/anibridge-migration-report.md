# Anibridge Migration — Verification Report

Date: 2026-06-05

## Summary

| Metric | Value |
|---|---|
| Distinct shows in gh-pages | 1,888 |
| Distinct shows in new source | 1,927 |
| Net change | +39 |
| Shows truly lost | 1 (remapped to correct TVDB ID) |
| Shows season-drifted | 7 (AniList reclassification) |
| Shows gained | 47 |
| Coverage improvement | ~2% higher distinct coverage |

## Lost shows detail

| TVDB ID | Title | Reason |
|---|---|---|
| 458472 | SHAMAN KING FLOWERS | Remapped to 383837 (mapping correction) |
| 462561 | Tenmaku no Jaadugar | Season drift — not in AniList summer 2026 anymore |
| 467841 | BanG Dream! Yume∞Mita | Season drift |
| 468399 | Hanaori-san Still Wants to Fight in the Next Life | Season drift |
| 470406 | Rich Girl Caretaker | Season drift |
| 474490 | "Kimi wo Aisuru Ki wa nai" | Season drift |
| 475631 | The Forsaken Saintess and Her Foodie Roadtrip | Season drift |
| 475980 | Mobius Dust | Season drift |

All 7 season-drifted shows are upcoming 2026 shows whose AniList season
classification changed since gh-pages was last generated. They are not
mapping losses — the new pipeline resolved whatever shows AniList is now
returning for each season. The 22 unmatched shows in summer 2026 represent
AniList entries not yet in the anibridge mapping (brand-new or obscure titles).

## Title changes

72 title changes across all files. These are cosmetic differences in how
TVDB or AniList formats titles (comma-separated multi-cour names, subtitle
order, romanization vs translated). No impact on Sonarr matching (which
uses `tvdbId`, not `title`).

## Reminder

The 7 season-drifted shows should be verified before summer 2026 airing.
See `docs/EXPLANATION.md` for the full list. If the anibridge dataset
still lacks these by then, consider adding the shinkro community mapping
as a fallback.

## Conclusion

**PASS** — deploy as the primary mapping source.
