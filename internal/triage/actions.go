package triage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Accept moves an item's media file (plus its .info.json/thumbnail/subtitle
// companions) out of Needs Review into organizedDir/<Artist>/, renamed to
// the canonical "Artist - Title" form, and records the human-confirmed
// artist/title back into the .info.json sidecar. It mirrors the move+rename
// logic in watcher.py's process_media so accepted files are indistinguishable
// from ones the postprocessor auto-organized.
func Accept(item Item, organizedDir, unknownArtist, unknownTitle, correctedArtist, correctedTitle string) (string, error) {
	artist := Sanitize(correctedArtist, unknownArtist)
	ext := filepath.Ext(item.mediaPath)
	canonical := BuildCanonicalTitle(artist, correctedTitle, unknownArtist, unknownTitle)

	destDir := filepath.Join(organizedDir, artist)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("creating artist dir: %w", err)
	}

	destMedia, err := ResolveCollision(filepath.Join(destDir, canonical+ext))
	if err != nil {
		return "", err
	}
	if err := moveFile(item.mediaPath, destMedia); err != nil {
		return "", fmt.Errorf("moving media: %w", err)
	}

	destStem := stemOf(destMedia)
	for _, companion := range item.companions {
		if err := moveCompanion(companion, destDir, destStem, item, artist, correctedTitle); err != nil {
			return "", fmt.Errorf("moving %s: %w", filepath.Base(companion), err)
		}
	}

	removeIfEmpty(filepath.Dir(item.mediaPath))
	return destMedia, nil
}

// Reject deletes an item's media file and all its companions — used when a
// download is junk and shouldn't be organized at all.
func Reject(item Item) error {
	if err := os.Remove(item.mediaPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, companion := range item.companions {
		if err := os.Remove(companion); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	removeIfEmpty(filepath.Dir(item.mediaPath))
	return nil
}

func moveCompanion(src, destDir, destStem string, item Item, confirmedArtist, confirmedTitle string) error {
	if src == item.infoPath {
		return moveInfoJSON(src, destDir, destStem, confirmedArtist, confirmedTitle)
	}

	// watcher.py normalizes thumbs/subs to dst_stem + their last suffix only
	// (e.g. both "Foo.webp" and "Foo.mp4.webp" become "<dst_stem>.webp") —
	// match that so re-triaged files look identical to auto-organized ones.
	suffix := strings.ToLower(filepath.Ext(src))
	dest, err := ResolveCollision(filepath.Join(destDir, destStem+suffix))
	if err != nil {
		return err
	}
	return moveFile(src, dest)
}

func moveInfoJSON(src, destDir, destStem, confirmedArtist, confirmedTitle string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	data["inferred_artist"] = confirmedArtist
	data["inferred_artist_confidence"] = 1.0
	data["inferred_title"] = confirmedTitle
	data["inferred_title_confidence"] = 1.0
	data["inferred_source"] = "human-triage"

	dest, err := ResolveCollision(filepath.Join(destDir, destStem+".info.json"))
	if err != nil {
		return err
	}
	if err := writeJSONAtomic(dest, data); err != nil {
		return err
	}
	return os.Remove(src)
}

func writeJSONAtomic(path string, data map[string]any) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// moveFile renames src to dst, falling back to copy+remove across devices
// (e.g. if Needs Review and Organized ever end up on different mounts).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}

func stemOf(path string) string {
	base := filepath.Base(path)
	return base[:len(base)-len(filepath.Ext(base))]
}

func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err == nil && len(entries) == 0 {
		os.Remove(dir)
	}
}
