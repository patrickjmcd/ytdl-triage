# ytdl-triage

Small web UI for manually resolving the low-confidence artist/title guesses
that [k8s-argo](https://github.com/patrickjmcd/k8s-argo)'s youtubedl
postprocessor (`apps/media/youtubedl/watcher.py`) couldn't make on its own.

When that pipeline can't infer an artist or title with enough confidence, it
drops the download (plus its `.info.json`, thumbnail, and subtitles) into a
`Needs Review/<guess>/` folder instead of organizing it automatically. This
app lists those items — with a thumbnail, the original metadata, and the
model's confidence scores — so you can correct the artist/title and move the
file into place, or delete it, from a browser instead of an SMB share.

## How it works

One Go binary serves both the JSON API (`/api/*`) and a built React/shadcn
SPA (embedded via `go:embed`, see `internal/web/`) — no separate frontend
deployment. There's no database: the filesystem is the source of truth, and
an item's ID is just a reversible encoding of its path relative to the Needs
Review directory.

- `GET /api/items` — scan `NEEDS_REVIEW_DIR` and list pending items
- `GET /api/items/{id}/thumbnail`, `GET /api/items/{id}/stream` — serve the
  thumbnail/video so you can preview before deciding
- `POST /api/items/{id}/accept` `{artist, title}` — move the media + sidecars
  into `ORGANIZED_DIR/<Artist>/`, renamed to the canonical `Artist - Title`
  form (naming logic ported from `watcher.py`'s `build_canonical_title`), and
  record the confirmed values back into the `.info.json`
- `DELETE /api/items/{id}` — delete the media + sidecars

## Configuration

Env vars (defaults match `watcher.py` so this can point at the same mount
with no extra config):

| Var | Default | |
|---|---|---|
| `PORT` | `8080` | |
| `ORGANIZED_DIR` | `/organized` | root the accepted file gets moved into |
| `NEEDS_REVIEW_DIR` | `$ORGANIZED_DIR/Needs Review` | |
| `UNKNOWN_ARTIST` | `Unknown Artist` | |
| `UNKNOWN_TITLE` | `Unknown Title` | |

## Development

```bash
# backend
go run ./cmd/server

# frontend (separate terminal; proxies /api to :8080)
cd frontend && npm install && npm run dev
```

Production-style single-binary run (mirrors the Dockerfile):

```bash
cd frontend && npm install && npm run build && cd ..
rm -rf internal/web/dist && cp -r frontend/dist internal/web/dist
go run ./cmd/server
```

`internal/web/dist/` is gitignored except for a placeholder `index.html` so
`go build` succeeds without a frontend build — see `go:embed` in
`internal/web/web.go`.

```bash
docker build -t ytdl-triage .
docker run -p 8080:8080 -v /path/to/organized:/organized ytdl-triage
```
