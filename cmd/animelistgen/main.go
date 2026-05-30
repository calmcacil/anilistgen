package main

import (
	"fmt"
	"os"

	"github.com/calmcacil/animelistgen/internal/config"
	"github.com/calmcacil/animelistgen/internal/season"
)

func main() {
	cfg := config.Parse()
	if cfg.Help {
		printUsage()
		os.Exit(0)
	}

	seasons := season.Resolve(cfg.Year, cfg.Season)
	if len(seasons) == 0 {
		fmt.Fprintln(os.Stderr, "no valid seasons to process")
		os.Exit(2)
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "animelistgen: processing %d season(s) for %d\n", len(seasons), cfg.Year)
	}

	if cfg.DryRun {
		for _, s := range seasons {
			fmt.Printf("[dry-run] would create list \"Anime %s %d\" with shows from AniList\n", s, cfg.Year)
		}
		return
	}

	if cfg.OutputDir != "" {
		// JSON output mode — TODO: implement
		fmt.Fprintf(os.Stderr, "output to directory: %s (not yet implemented)\n", cfg.OutputDir)
		return
	}

	if cfg.MDBListKey == "" {
		fmt.Fprintln(os.Stderr, "MDBList API key required. Set -key flag or MDBLIST_API_KEY env var.")
		os.Exit(1)
	}

	// TODO: implement
	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(0)
}

func printUsage() {
	fmt.Println(`animelistgen — generate seasonal anime lists on MDBList from AniList

Usage:
  animelistgen [flags]

Flags:
  -y, -year INT       target year (default: current year)
  -s, -season STR     single season: winter|spring|summer|fall (default: all)
  -k, -mdblist-key STR  MDBList API key (or MDBLIST_API_KEY env var)
  -o, -output DIR     write JSON files to DIR instead of calling MDBList
  -dry-run            print what would be done without calling MDBList
  -v, -verbose        print progress to stderr
  -h, -help           print this help`)
}
