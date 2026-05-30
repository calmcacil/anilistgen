package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/calmcacil/anilistgen/internal/anilist"
	"github.com/calmcacil/anilistgen/internal/mdblist"
)

// ResolvedListTitle fills in placeholders in a title template.
func ResolvedListTitle(template, season string, year int) string {
	s := strings.ReplaceAll(template, "{season}", capitalize(season))
	s = strings.ReplaceAll(s, "{year}", fmt.Sprintf("%d", year))
	return s
}

// ResolvedListDescription fills in placeholders in a description template.
func ResolvedListDescription(template, season string, year int) string {
	s := strings.ReplaceAll(template, "{season}", capitalize(season))
	s = strings.ReplaceAll(s, "{year}", fmt.Sprintf("%d", year))
	return s
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// mdbItem represents a resolved item for adding to MDBList.
type mdbItem struct {
	id    map[string]any // provider ID for MDBList API
	title string        // display title
}

// Result holds the outcome of syncing one season.
type Result struct {
	Season           string
	Year             int
	ListTitle        string
	ListURL          string
	ShowCount        int
	TotalInDB        int // shows found in MDBList's database
	FoundViaFallback int // shows matched via relation fallback
	Skipped          int // shows skipped (not in MDBList)
	SkippedDuration  int // shows skipped (duration <= 10 min)
	SkippedBlacklist int // shows skipped (blacklisted)
	Created          bool
	Updated          bool
	Error            error
}

// SeasonResult holds the output for a single season sync (JSON output mode).
type SeasonResult struct {
	Season    string       `json:"season"`
	Year      int          `json:"year"`
	Timestamp string       `json:"timestamp"`
	Shows     []anilist.Show `json:"shows"`
}

// Syncer orchestrates fetching from AniList and publishing to MDBList.
type Syncer struct {
	anilist *anilist.Client
	mdblist *mdblist.Client
	cfg     SyncConfig
}

// SyncConfig holds the parameters for a sync operation.
type SyncConfig struct {
	MaxPerSeason            int
	IncludeONA              bool
	TitleTemplate           string
	DescriptionTemplate     string
	Public                  bool
	DryRun                  bool
	OutputDir               string
	Blacklist               []string
	FallbackRelationTypes   []string
}

// isBlacklisted checks if a show should be skipped.
func (c *SyncConfig) isBlacklisted(title string, idMal int) bool {
	for _, entry := range c.Blacklist {
		if entry == "" {
			continue
		}
		// Numeric entry → match against MAL ID
		var malID int
		if _, err := fmt.Sscanf(entry, "%d", &malID); err == nil && malID > 0 {
			if malID == idMal {
				return true
			}
			continue
		}
		// Text entry → case-insensitive substring match against title
		if strings.Contains(strings.ToLower(title), strings.ToLower(entry)) {
			return true
		}
	}
	return false
}

// New creates a new Syncer.
func New(ani *anilist.Client, mdb *mdblist.Client, cfg SyncConfig) *Syncer {
	return &Syncer{
		anilist: ani,
		mdblist: mdb,
		cfg:     cfg,
	}
}

// SyncSeason fetches anime for a single season and publishes to MDBList.
func (s *Syncer) SyncSeason(ctx context.Context, season string, year int) Result {
	title := ResolvedListTitle(s.cfg.TitleTemplate, season, year)
	desc := ResolvedListDescription(s.cfg.DescriptionTemplate, season, year)

	slog.Debug("fetching season", "season", season, "year", year)

	shows, err := s.anilist.FetchSeason(ctx, season, year, s.cfg.MaxPerSeason, s.cfg.IncludeONA)
	if err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("fetch AniList: %w", err),
		}
	}

	if len(shows) >= s.cfg.MaxPerSeason && s.cfg.MaxPerSeason > 0 {
		slog.Warn("season may have more results than max_per_season",
			"season", season, "year", year, "got", len(shows), "max", s.cfg.MaxPerSeason)
	}

	// Filter out shows with duration <= 10 minutes (shorts, music videos, etc.)
	var filtered []anilist.Show
	var skippedDuration, skippedBlacklist int
	for _, show := range shows {
		title := show.DisplayTitle()
		idMal := 0
		if show.IDMal != nil {
			idMal = *show.IDMal
		}

		if show.SkipByDuration() {
			skippedDuration++
			slog.Debug("skipped show (duration <= 10 min)",
				"title", title,
				"duration", show.Duration)
			continue
		}

		if s.cfg.isBlacklisted(title, idMal) {
			skippedBlacklist++
			slog.Debug("skipped show (blacklisted)",
				"title", title,
				"idMal", idMal)
			continue
		}

		filtered = append(filtered, show)
	}
	shows = filtered

	totalSkipped := skippedDuration + skippedBlacklist
	if totalSkipped > 0 {
		slog.Info("filtered shows",
			"season", season, "year", year,
			"skipped_duration", skippedDuration,
			"skipped_blacklist", skippedBlacklist,
			"remaining", len(shows))
	}

	// Output-to-file mode
	if s.cfg.OutputDir != "" {
		return s.writeJSONOutput(season, year, shows, title)
	}

	// Dry-run mode
	if s.cfg.DryRun {
		slog.Info("dry-run: would process list",
			"title", title,
			"season", season,
			"year", year,
			"shows", len(shows))
		return Result{
			Season:    season,
			Year:      year,
			ListTitle: title,
			ShowCount: len(shows),
			Skipped:   skippedDuration,
		}
	}

	// Normal mode: find/create/update MDBList
	result := s.syncMDBList(ctx, season, year, title, desc, shows)
	// Merge filter stats
	result.SkippedDuration = skippedDuration
	result.SkippedBlacklist = skippedBlacklist
	result.Skipped += totalSkipped
	return result
}

// writeJSONOutput writes show data to a JSON file.
func (s *Syncer) writeJSONOutput(season string, year int, shows []anilist.Show, title string) Result {
	result := SeasonResult{
		Season:    season,
		Year:      year,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Shows:     shows,
	}

	filename := fmt.Sprintf("anime-%s-%d.json", strings.ToLower(season), year)
	outPath := filepath.Join(s.cfg.OutputDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("marshal JSON: %w", err),
		}
	}

	if err := os.MkdirAll(s.cfg.OutputDir, 0755); err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("create output dir: %w", err),
		}
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("write JSON file: %w", err),
		}
	}

	slog.Debug("wrote JSON output", "file", outPath, "shows", len(shows))

	return Result{
		Season:    season,
		Year:      year,
		ListTitle: title,
		ShowCount: len(shows),
		Skipped:   0,
	}
}

// syncMDBList does the actual MDBList list creation/update with items.
func (s *Syncer) syncMDBList(ctx context.Context, season string, year int, title, desc string, shows []anilist.Show) Result {
	slog.Debug("looking up existing list", "title", title)

	existing, err := s.mdblist.FindListByTitle(ctx, title)
	if err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("find existing list: %w", err),
		}
	}

	// Collect all MAL IDs: direct IDs + relation (prequel) IDs for fallback
	type showItem struct {
		show       anilist.Show
		directMAL  int  // show's own MAL ID (0 if none)
		fallbackID int  // resolved fallback MAL ID (0 if none)
		found      bool // whether a match was found in MDBList
	}

	items := make([]showItem, len(shows))
	allMALIDs := make(map[int]bool) // dedup set
	for i, sh := range shows {
		items[i] = showItem{show: sh}
		if sh.IDMal != nil && *sh.IDMal > 0 {
			items[i].directMAL = *sh.IDMal
			allMALIDs[*sh.IDMal] = true
		}
		// Also collect fallback relation MAL IDs (e.g. PREQUEL, PARENT)
		// based on the configured FallbackRelationTypes.
		for _, relID := range sh.RelationMALIDsByType(s.cfg.FallbackRelationTypes) {
			if relID > 0 && relID != items[i].directMAL {
				allMALIDs[relID] = true
			}
		}
	}

	// Batch lookup ALL MAL IDs (direct + relations) in one call
	allIDs := make([]int, 0, len(allMALIDs))
	for id := range allMALIDs {
		allIDs = append(allIDs, id)
	}

	malInfoMap := map[int]mdblist.MediaInfo{}
	if len(allIDs) > 0 && s.mdblist != nil {
		var lookupErr error
		malInfoMap, lookupErr = s.mdblist.BatchLookupByMAL(ctx, allIDs)
		if lookupErr != nil {
			slog.Warn("MAL batch lookup failed", "error", lookupErr)
			malInfoMap = map[int]mdblist.MediaInfo{}
		}
	}

	// Build items for MDBList, trying fallback if direct match fails
	var mdbItems []mdbItem
	var foundDirect, foundFallback, notFoundCount int

	for i := range items {
		it := items[i]
		displayTitle := it.show.DisplayTitle()

		// Try direct MAL ID first
		if it.directMAL > 0 {
			if info, ok := malInfoMap[it.directMAL]; ok {
				id := map[string]any{}
				if info.IDs.IMDB != "" {
					id["imdb"] = info.IDs.IMDB
				} else if info.IDs.TMDB != 0 {
					id["tmdb"] = info.IDs.TMDB
				} else if info.IDs.TVDB != 0 {
					id["tvdb"] = info.IDs.TVDB
				}
				mdbItems = append(mdbItems, mdbItem{id: id, title: displayTitle})
				items[i].found = true
				foundDirect++
				continue
			}
		}

		// Direct MAL ID not found — try fallback relation MAL IDs
		// based on configured FallbackRelationTypes.
		for _, relID := range it.show.RelationMALIDsByType(s.cfg.FallbackRelationTypes) {
			if relID == it.directMAL {
				continue
			}
			if info, ok := malInfoMap[relID]; ok {
				id := map[string]any{}
				if info.IDs.IMDB != "" {
					id["imdb"] = info.IDs.IMDB
				} else if info.IDs.TMDB != 0 {
					id["tmdb"] = info.IDs.TMDB
				} else if info.IDs.TVDB != 0 {
					id["tvdb"] = info.IDs.TVDB
				}
				mdbItems = append(mdbItems, mdbItem{id: id, title: displayTitle})
				items[i].found = true
				items[i].fallbackID = relID
				foundFallback++
				slog.Debug("matched via relation fallback",
					"title", displayTitle,
					"directMAL", it.directMAL,
					"fallbackMAL", relID,
					"fallbackTitle", info.Title)
				break
			}
		}

		// Show not found in MDBList at all — skip
		notFoundCount++
		slog.Debug("show not in MDBList, skipping",
			"title", displayTitle,
			"idMal", it.directMAL)
	}

	if notFoundCount > 0 {
		slog.Info("some shows not found in MDBList database",
			"season", season, "year", year,
			"not_found", notFoundCount,
			"direct_matches", foundDirect,
			"fallback_matches", foundFallback,
			"total", len(shows))
	}

	if existing != nil {
		slog.Debug("found existing list", "id", existing.ID, "title", title)

		// Recreate with new items (delete + create with items)
		slog.Info("replacing list items",
			"title", title,
			"id", existing.ID,
			"items", len(mdbItems))

		newList, err := s.mdblist.DeleteAndRecreate(ctx, existing.ID, title, desc, s.cfg.Public, providerIDs(mdbItems))
		if err != nil {
			return Result{
				Season: season,
				Year:   year,
				Error:  fmt.Errorf("replace list: %w", err),
			}
		}

		return Result{
			Season:           season,
			Year:             year,
			ListTitle:        title,
			ListURL:          newList.GetURL(),
			ShowCount:        len(shows),
			TotalInDB:        foundDirect + foundFallback,
			FoundViaFallback: foundFallback,
			Skipped:          notFoundCount,
			Updated:          true,
		}
	}

	// Create new list, then add items
	slog.Info("creating new list",
		"title", title,
		"season", season,
		"year", year,
		"items", len(mdbItems))

	newList, err := s.mdblist.CreateList(ctx, title, desc, s.cfg.Public)
	if err != nil {
		return Result{
			Season: season,
			Year:   year,
			Error:  fmt.Errorf("create list: %w", err),
		}
	}

	// Add items in batches
	ids := providerIDs(mdbItems)
	if len(ids) > 0 {
		const batchSize = 200
		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			if err := s.mdblist.AddItems(ctx, newList.ID, ids[i:end]); err != nil {
				return Result{
					Season: season,
					Year:   year,
					Error:  fmt.Errorf("add items: %w", err),
				}
			}
		}
	}

	return Result{
		Season:           season,
		Year:             year,
		ListTitle:        title,
		ListURL:          newList.GetURL(),
		ShowCount:        len(shows),
		TotalInDB:        foundDirect + foundFallback,
		FoundViaFallback: foundFallback,
		Skipped:          notFoundCount,
		Created:          true,
	}
}

// providerIDs extracts the provider ID maps from mdbItem slice.
func providerIDs(items []mdbItem) []map[string]any {
	ids := make([]map[string]any, len(items))
	for i, it := range items {
		ids[i] = it.id
	}
	return ids
}

// SyncAll processes all configured seasons.
func (s *Syncer) SyncAll(ctx context.Context, seasons []string, year int) []Result {
	results := make([]Result, 0, len(seasons))
	for _, season := range seasons {
		r := s.SyncSeason(ctx, season, year)
		results = append(results, r)
	}
	return results
}

// PrintResults prints sync results in a human-readable format.
func PrintResults(results []Result, dryRun bool) {
	for _, r := range results {
		if r.Error != nil {
			slog.Error("sync failed",
				"season", r.Season,
				"year", r.Year,
				"error", r.Error)
			continue
		}

		var parts []string

		if r.ShowCount > 0 {
			parts = append(parts, fmt.Sprintf("%d shows", r.ShowCount))
			if r.TotalInDB > 0 {
				if r.FoundViaFallback > 0 {
					parts = append(parts, fmt.Sprintf("%d via fallback", r.FoundViaFallback))
				}
				if r.TotalInDB < r.ShowCount {
					parts = append(parts, fmt.Sprintf("%d in MDBList", r.TotalInDB))
				}
			}
		} else {
			parts = append(parts, "no shows")
		}

		var skippedParts []string
		if r.SkippedDuration > 0 {
			skippedParts = append(skippedParts, fmt.Sprintf("%d short", r.SkippedDuration))
		}
		if r.SkippedBlacklist > 0 {
			skippedParts = append(skippedParts, fmt.Sprintf("%d blacklisted", r.SkippedBlacklist))
		}
		if r.Skipped > 0 {
			skippedParts = append(skippedParts, fmt.Sprintf("%d not in MDB", r.Skipped))
		}
		if len(skippedParts) > 0 {
			skippedStr := "skipped: " + strings.Join(skippedParts, ", ")
			parts = append(parts, skippedStr)
		}

		if dryRun || (r.Created == false && r.Updated == false) {
			detail := strings.Join(parts, ", ")
			fmt.Printf("[dry-run] %s %d: %s — %s\n",
				r.Season, r.Year, r.ListTitle, detail)
			continue
		}

		status := "created"
		if r.Updated {
			status = "updated"
		}

		detail := strings.Join(parts, ", ")

		if r.ListURL != "" {
			fmt.Printf("%s %d: %s — %s — %s — %s\n",
				r.Season, r.Year, r.ListTitle, status, detail, r.ListURL)
		} else {
			fmt.Printf("%s %d: %s — %s — %s\n",
				r.Season, r.Year, r.ListTitle, status, detail)
		}
	}
}
