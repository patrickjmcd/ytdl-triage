package triage

import "strings"

// Mirrors watcher.py's classify_category — used here purely to show the
// same category label a human reviewer would otherwise see once the file
// reaches Organized/, not to filter anything.
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
