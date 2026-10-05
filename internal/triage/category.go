package triage

import "strings"

// Mirrors watcher.py's classify_category — used here purely to show the
// same category label a human reviewer would otherwise see once the file
// reaches Organized/, not to filter anything.
//
// DUPLICATION RISK: this is a hand-ported copy, not a shared dependency —
// watcher.py lives in k8s-argo (apps/media/youtubedl/watcher.py) and this
// repo has no code link to it. If the hint lists or category rules change
// there, update this function (and naming.go's Sanitize/BuildCanonicalTitle/
// ResolveCollision, also ported from watcher.py) to match, or the two will
// silently drift.
var (
	tinyDeskHints        = []string{"tiny desk"}
	livePerformanceHints = []string{
		"full concert", "full set", "full show", "complete", "entire", "livestream",
		"festival", "glastonbury", "lollapalooza", "outside lands", "acl",
		"live at", "session", "interview", "performance + interview",
	}
)

func classifyCategory(title, desc string) string {
	t := strings.ToLower(title)
	d := strings.ToLower(desc)
	for _, h := range tinyDeskHints {
		if strings.Contains(t, h) || strings.Contains(d, h) {
			return "Tiny Desk Concert"
		}
	}
	for _, h := range livePerformanceHints {
		if strings.Contains(t, h) || strings.Contains(d, h) {
			return "Live Performance"
		}
	}
	return "Music Video"
}
