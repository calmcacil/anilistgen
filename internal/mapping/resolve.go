package mapping

import (
	"log/slog"

	"github.com/calmcacil/anilistgen/internal/model"
	"github.com/calmcacil/anilistgen/internal/output"
)

type Resolver struct {
	source *AnibridgeMapping
}

func NewResolver(am *AnibridgeMapping) *Resolver {
	return &Resolver{source: am}
}

func (r *Resolver) Project(shows []model.Show) []output.Show {
	var result []output.Show
	for _, show := range shows {
		tvdbID, ok := r.lookup(show)
		if !ok {
			continue
		}
		slog.Debug("resolved via anibridge",
			"title", show.DisplayTitle(),
			"anilist", show.ID, "mal", deref(show.IDMal),
			"tvdb", tvdbID)
		result = append(result, output.Show{
			TVDBID: tvdbID,
			Title:  show.DisplayTitle(),
		})
	}
	return result
}

func (r *Resolver) lookup(s model.Show) (int, bool) {
	if s.IDMal != nil && *s.IDMal > 0 {
		if id, ok := r.source.LookupByMAL(*s.IDMal); ok {
			return id, true
		}
	}
	if s.ID > 0 {
		if id, ok := r.source.LookupByAniList(s.ID); ok {
			return id, true
		}
	}
	return 0, false
}

func deref[T any](p *T) T {
	var zero T
	if p != nil {
		return *p
	}
	return zero
}
