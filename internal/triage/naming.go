package triage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DUPLICATION RISK: every function in this file is a hand-ported copy of
// logic in watcher.py (k8s-argo's apps/media/youtubedl/watcher.py), not a
// shared dependency — there's no code link between the two repos. If the
// naming/collision rules change there, update Sanitize/BuildCanonicalTitle/
// ResolveCollision here (and classifyCategory in category.go) to match, or
// accepted files will end up named differently than ones watcher.py
// auto-organized.
var (
	illegalCharsRe = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	whitespaceRe   = regexp.MustCompile(`\s+`)
	andSplitRe     = regexp.MustCompile(`(?i)\s*&\s*|\s+and\s+`)
)

// Sanitize strips filesystem-illegal characters and collapses whitespace,
// matching watcher.py's sanitize() exactly so renamed files stay consistent
// with ones the postprocessor named itself.
func Sanitize(s, fallback string) string {
	s = strings.TrimSpace(s)
	s = illegalCharsRe.ReplaceAllString(s, "_")
	s = whitespaceRe.ReplaceAllString(s, " ")
	s = strings.TrimRight(s, " .")
	if s == "" {
		return fallback
	}
	return s
}

// artistLeadPattern builds a regex matching a leading occurrence of artist at
// the start of a string, treating "&" and "and" as interchangeable the same
// way watcher.py's _artist_lead_pattern does.
func artistLeadPattern(artist string) *regexp.Regexp {
	parts := andSplitRe.Split(artist, -1)
	var escaped []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			escaped = append(escaped, regexp.QuoteMeta(p))
		}
	}
	if len(escaped) == 0 {
		return nil
	}
	body := strings.Join(escaped, `\s*(?:&|and)\s*`)
	return regexp.MustCompile(`(?i)^\s*` + body + `\s*[-:]*\s*`)
}

// BuildCanonicalTitle prefixes title with artist unless a single leading
// mention already exists, collapsing any genuine repeat into exactly one
// canonical "Artist - Title" — a direct port of watcher.py's
// build_canonical_title so manually-corrected filenames match the
// postprocessor's own naming convention.
func BuildCanonicalTitle(artist, title, unknownArtist, fallback string) string {
	if artist == unknownArtist {
		return Sanitize(title, fallback)
	}

	pat := artistLeadPattern(artist)
	if pat == nil {
		return Sanitize(title, fallback)
	}

	loc := pat.FindStringIndex(title)
	if loc == nil {
		return Sanitize(fmt.Sprintf("%s - %s", artist, title), fallback)
	}

	rest := title[loc[1]:]
	if !pat.MatchString(rest) {
		// Exactly one mention already, however it's formatted — leave it alone.
		return Sanitize(title, fallback)
	}

	// A genuine repeat: strip every further leading repeat, then rebuild once.
	prev := ""
	for rest != prev {
		prev = rest
		nextLoc := pat.FindStringIndex(rest)
		if nextLoc != nil && nextLoc[1] > 0 {
			candidate := strings.TrimSpace(rest[nextLoc[1]:])
			if candidate != "" {
				rest = candidate
			}
		}
	}
	return Sanitize(fmt.Sprintf("%s - %s", artist, rest), fallback)
}

// ResolveCollision appends " (2)", " (3)", ... before the extension until it
// finds a path that doesn't already exist, matching watcher.py's
// resolve_collision.
func ResolveCollision(dst string) (string, error) {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return dst, nil
	}

	ext := filepath.Ext(dst)
	base := strings.TrimSuffix(filepath.Base(dst), ext)
	dir := filepath.Dir(dst)

	for i := 2; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("too many collisions for %s", dst)
}
