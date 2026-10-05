package triage

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Scan walks needsReviewDir one level deep — <ArtistGuess>/<file> — the exact
// layout watcher.py creates when it drops a low-confidence download there —
// and returns one Item per media file found, newest first.
func Scan(needsReviewDir string) ([]Item, error) {
	artistDirs, err := os.ReadDir(needsReviewDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, err
	}

	items := []Item{}
	for _, ad := range artistDirs {
		if !ad.IsDir() {
			continue
		}
		artistDir := filepath.Join(needsReviewDir, ad.Name())
		files, err := os.ReadDir(artistDir)
		if err != nil {
			continue
		}

		names := make(map[string]os.DirEntry, len(files))
		for _, f := range files {
			names[f.Name()] = f
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if !MediaExts[ext] || SkipExts[ext] {
				continue
			}

			mediaPath := filepath.Join(artistDir, f.Name())
			item, err := buildItem(needsReviewDir, artistDir, ad.Name(), f, names)
			if err != nil {
				continue
			}
			_ = mediaPath
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModTime.After(items[j].ModTime)
	})
	return items, nil
}

func buildItem(needsReviewDir, artistDir, artistGuess string, f os.DirEntry, siblings map[string]os.DirEntry) (Item, error) {
	info, err := f.Info()
	if err != nil {
		return Item{}, err
	}

	mediaPath := filepath.Join(artistDir, f.Name())
	stem := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

	id := encodeID(needsReviewDir, mediaPath)

	item := Item{
		ID:          id,
		ArtistGuess: artistGuess,
		Filename:    f.Name(),
		SizeBytes:   info.Size(),
		ModTime:     info.ModTime(),
		mediaPath:   mediaPath,
	}

	// yt-dlp's default sidecar is Foo.mp4.info.json; some setups produce
	// Foo.info.json instead — check both, same as find_info_json().
	if _, ok := siblings[f.Name()+".info.json"]; ok {
		item.infoPath = filepath.Join(artistDir, f.Name()+".info.json")
	} else if _, ok := siblings[stem+".info.json"]; ok {
		item.infoPath = filepath.Join(artistDir, stem+".info.json")
	}

	for ext := range ThumbExts {
		if _, ok := siblings[stem+ext]; ok {
			item.thumbPath = filepath.Join(artistDir, stem+ext)
			item.HasThumbnail = true
			break
		}
		if _, ok := siblings[f.Name()+ext]; ok {
			item.thumbPath = filepath.Join(artistDir, f.Name()+ext)
			item.HasThumbnail = true
			break
		}
	}

	for ext := range SubExts {
		if _, ok := siblings[stem+ext]; ok {
			item.companions = append(item.companions, filepath.Join(artistDir, stem+ext))
		}
		if _, ok := siblings[f.Name()+ext]; ok {
			item.companions = append(item.companions, filepath.Join(artistDir, f.Name()+ext))
		}
	}
	if item.infoPath != "" {
		item.companions = append(item.companions, item.infoPath)
	}
	if item.thumbPath != "" {
		item.companions = append(item.companions, item.thumbPath)
	}

	item.RawTitle = stem
	if item.infoPath != "" {
		if raw, err := os.ReadFile(item.infoPath); err == nil {
			var data map[string]any
			if json.Unmarshal(raw, &data) == nil {
				applyInfoJSON(&item, data)
			}
		}
	}

	return item, nil
}

func applyInfoJSON(item *Item, data map[string]any) {
	if v := strString(data["title"]); v != "" {
		item.RawTitle = v
	} else if v := strString(data["fulltitle"]); v != "" {
		item.RawTitle = v
	}
	item.Description = strString(data["description"])
	if v := strString(data["uploader"]); v != "" {
		item.Uploader = v
	} else {
		item.Uploader = strString(data["channel"])
	}

	item.InferredArtist = strString(data["inferred_artist"])
	item.InferredArtistConfidence = floatVal(data["inferred_artist_confidence"])
	item.InferredTitle = strString(data["inferred_title"])
	item.InferredTitleConfidence = floatVal(data["inferred_title_confidence"])
	item.InferredIsFullSet, _ = data["inferred_is_full_set"].(bool)
	item.InferredSource = strString(data["inferred_source"])
	item.Category = classifyCategory(item.RawTitle, item.Description)
}

func strString(v any) string {
	s, _ := v.(string)
	return s
}

func floatVal(v any) float64 {
	f, _ := v.(float64)
	return f
}

// encodeID turns a media file's path (relative to needsReviewDir) into a
// URL-safe, reversible item ID — no database needed, the filesystem is the
// source of truth.
func encodeID(needsReviewDir, mediaPath string) string {
	rel, err := filepath.Rel(needsReviewDir, mediaPath)
	if err != nil {
		rel = mediaPath
	}
	return base64.RawURLEncoding.EncodeToString([]byte(rel))
}

// FindItem resolves a single item by ID without re-scanning the whole
// directory tree — used by the accept/reject handlers.
func FindItem(needsReviewDir, id string) (Item, error) {
	mediaPath, err := DecodeID(needsReviewDir, id)
	if err != nil {
		return Item{}, err
	}

	artistDir := filepath.Dir(mediaPath)
	artistGuess, err := filepath.Rel(needsReviewDir, artistDir)
	if err != nil {
		return Item{}, err
	}

	siblingEntries, err := os.ReadDir(artistDir)
	if err != nil {
		return Item{}, err
	}
	siblings := make(map[string]os.DirEntry, len(siblingEntries))
	var self os.DirEntry
	for _, f := range siblingEntries {
		siblings[f.Name()] = f
		if f.Name() == filepath.Base(mediaPath) {
			self = f
		}
	}
	if self == nil {
		return Item{}, os.ErrNotExist
	}

	return buildItem(needsReviewDir, artistDir, artistGuess, self, siblings)
}

// DecodeID resolves an item ID back to an absolute path, rejecting anything
// that would escape needsReviewDir (defense against a crafted ID).
func DecodeID(needsReviewDir, id string) (string, error) {
	relBytes, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", err
	}
	rel := filepath.Clean(string(relBytes))
	if rel == ".." || strings.HasPrefix(rel, "../") || strings.HasPrefix(rel, "/") {
		return "", os.ErrInvalid
	}
	abs := filepath.Join(needsReviewDir, rel)
	if !strings.HasPrefix(abs, filepath.Clean(needsReviewDir)+string(filepath.Separator)) {
		return "", os.ErrInvalid
	}
	return abs, nil
}
