package triage

import "time"

var (
	MediaExts = map[string]bool{".mp4": true, ".mkv": true, ".webm": true, ".mov": true, ".m4v": true}
	SkipExts  = map[string]bool{".part": true, ".tmp": true, ".ytdl": true, ".download": true, ".crdownload": true}
	ThumbExts = map[string]bool{".webp": true, ".jpg": true, ".jpeg": true, ".png": true}
	SubExts   = map[string]bool{".vtt": true, ".srt": true, ".ass": true, ".lrc": true}
)

// Item is one media file sitting in the Needs Review directory, bundled with
// whatever sidecars (.info.json, thumbnail, subtitles) share its filename
// stem — the same grouping watcher.py uses when it first moved the file here.
type Item struct {
	ID                       string    `json:"id"`
	ArtistGuess              string    `json:"artistGuess"`
	Filename                 string    `json:"filename"`
	SizeBytes                int64     `json:"sizeBytes"`
	ModTime                  time.Time `json:"modTime"`
	RawTitle                 string    `json:"rawTitle"`
	Description              string    `json:"description"`
	Uploader                 string    `json:"uploader"`
	Category                 string    `json:"category"`
	InferredArtist           string    `json:"inferredArtist"`
	InferredArtistConfidence float64   `json:"inferredArtistConfidence"`
	InferredTitle            string    `json:"inferredTitle"`
	InferredTitleConfidence  float64   `json:"inferredTitleConfidence"`
	InferredIsFullSet        bool      `json:"inferredIsFullSet"`
	InferredSource           string    `json:"inferredSource"`
	HasThumbnail             bool      `json:"hasThumbnail"`

	// mediaPath is the absolute on-disk path to the media file. Not exposed
	// in JSON; handlers resolve it fresh from ID on every request instead of
	// trusting a cached value, since another triage pass may have moved it.
	mediaPath  string   `json:"-"`
	infoPath   string   `json:"-"`
	thumbPath  string   `json:"-"`
	companions []string `json:"-"`
}

func (i Item) MediaPath() string { return i.mediaPath }
func (i Item) ThumbPath() string { return i.thumbPath }
