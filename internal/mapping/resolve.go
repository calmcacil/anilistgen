package mapping

import (
	"context"
	"log/slog"

	"github.com/calmcacil/anilistgen/internal/anilist"
)

type Resolver struct {
	community  *CommunityMapping
	animelists *AnimeListsMapping
}

type ResolvedShow struct {
	MALID    int
	TVDBID   int
	Title    string
	Resolved bool
}

func NewResolver(cm *CommunityMapping, alm *AnimeListsMapping) *Resolver {
	return &Resolver{
		community:  cm,
		animelists: alm,
	}
}

func (r *Resolver) Resolve(ctx context.Context, malID int, anilistID int, title string) (int, bool) {
	if malID <= 0 {
		return 0, false
	}

	if anilistID > 0 {
		if tvdbID, ok := r.animelists.Lookup(anilistID); ok {
			slog.Debug("resolved via anime-lists",
				"title", title, "mal", malID, "anilist_id", anilistID, "tvdb", tvdbID)
			return tvdbID, true
		}
	}

	if tvdbID, ok := r.community.Lookup(malID); ok {
		slog.Debug("resolved via community mapping",
			"title", title, "mal", malID, "tvdb", tvdbID)
		return tvdbID, true
	}

	return 0, false
}

func (r *Resolver) ResolveBatch(ctx context.Context, shows []anilist.Show) map[int]ResolvedShow {
	result := make(map[int]ResolvedShow, len(shows))

	for _, show := range shows {
		malID := 0
		if show.IDMal != nil {
			malID = *show.IDMal
		}

		rs := ResolvedShow{
			MALID: malID,
			Title: show.DisplayTitle(),
		}

		if tvdbID, ok := r.Resolve(ctx, malID, show.ID, rs.Title); ok {
			rs.TVDBID = tvdbID
			rs.Resolved = true
		}

		result[show.ID] = rs
	}

	return result
}
