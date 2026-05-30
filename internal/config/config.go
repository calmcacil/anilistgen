package config

import (
	"flag"
	"os"
	"time"
)

// Config holds all CLI flags and derived configuration.
type Config struct {
	Year       int
	Season     string
	MDBListKey string
	OutputDir  string
	DryRun     bool
	Verbose    bool
	Help       bool
}

// Parse reads CLI flags and environment variables, returning a Config.
func Parse() *Config {
	var cfg Config

	// Custom flag usage
	flag.Usage = func() {
		// usage is printed by main on -h
	}

	flag.IntVar(&cfg.Year, "y", time.Now().Year(), "target year")
	flag.IntVar(&cfg.Year, "year", time.Now().Year(), "target year")
	flag.StringVar(&cfg.Season, "s", "", "single season: winter|spring|summer|fall")
	flag.StringVar(&cfg.Season, "season", "", "single season: winter|spring|summer|fall")
	flag.StringVar(&cfg.MDBListKey, "k", "", "MDBList API key")
	flag.StringVar(&cfg.MDBListKey, "mdblist-key", "", "MDBList API key")
	flag.StringVar(&cfg.OutputDir, "o", "", "write JSON files to directory")
	flag.StringVar(&cfg.OutputDir, "output", "", "write JSON files to directory")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "dry run mode")
	flag.BoolVar(&cfg.Verbose, "v", false, "verbose output")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "verbose output")
	flag.BoolVar(&cfg.Help, "h", false, "print help")
	flag.BoolVar(&cfg.Help, "help", false, "print help")

	flag.Parse()

	// Fallback: env var for MDBList key
	if cfg.MDBListKey == "" {
		cfg.MDBListKey = os.Getenv("MDBLIST_API_KEY")
	}

	return &cfg
}
