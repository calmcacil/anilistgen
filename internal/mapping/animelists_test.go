package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnimeListsMapping_Lookup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "anime-list-full.xml")
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<completelist>
<anime anidbid="12345" tvdbid="78901" defaulttvdbseason="1" />
<anime anidbid="67890" tvdbid="23456" defaulttvdbseason="" />
</completelist>`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	alm, err := LoadAnimeListsMapping(path)
	if err != nil {
		t.Fatalf("LoadAnimeListsMapping: %v", err)
	}

	tvdb, ok := alm.Lookup(12345)
	if !ok {
		t.Error("expected to find AniDB 12345")
	}
	if tvdb != 78901 {
		t.Errorf("expected TVDB 78901, got %d", tvdb)
	}

	tvdb, ok = alm.Lookup(67890)
	if !ok {
		t.Error("expected to find AniDB 67890")
	}
	if tvdb != 23456 {
		t.Errorf("expected TVDB 23456, got %d", tvdb)
	}
}

func TestAnimeListsMapping_Lookup_Missing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "anime-list-full.xml")
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<completelist>
<anime anidbid="12345" tvdbid="78901" />
</completelist>`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	alm, err := LoadAnimeListsMapping(path)
	if err != nil {
		t.Fatalf("LoadAnimeListsMapping: %v", err)
	}

	_, ok := alm.Lookup(99999)
	if ok {
		t.Error("expected false for missing AniDB ID")
	}
}

func TestAnimeListsMapping_SkipsInvalidEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "anime-list-full.xml")
	content := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<completelist>
<anime anidbid="0" tvdbid="78901" />
<anime anidbid="abc" tvdbid="78901" />
<anime anidbid="12345" tvdbid="0" />
<anime anidbid="12345" tvdbid="abc" />
<anime anidbid="67890" tvdbid="23456" />
</completelist>`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	alm, err := LoadAnimeListsMapping(path)
	if err != nil {
		t.Fatalf("LoadAnimeListsMapping: %v", err)
	}

	if len(alm.data) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(alm.data))
	}
}

func TestAnimeListsMapping_InvalidXML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.xml")
	content := []byte(`not xml at all`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAnimeListsMapping(path)
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}
