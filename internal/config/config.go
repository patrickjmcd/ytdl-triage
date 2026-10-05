// Package config loads ytdl-triage's settings from environment variables.
// Defaults mirror the youtubedl postprocessor's watcher.py so this app can be
// pointed at the same ORGANIZED_DIR mount with no extra configuration.
package config

import "os"

type Config struct {
	Port           string
	OrganizedDir   string
	NeedsReviewDir string
	UnknownArtist  string
	UnknownTitle   string
}

func Load() Config {
	organized := getenv("ORGANIZED_DIR", "/organized")
	return Config{
		Port:           getenv("PORT", "8080"),
		OrganizedDir:   organized,
		NeedsReviewDir: getenv("NEEDS_REVIEW_DIR", organized+"/Needs Review"),
		UnknownArtist:  getenv("UNKNOWN_ARTIST", "Unknown Artist"),
		UnknownTitle:   getenv("UNKNOWN_TITLE", "Unknown Title"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
