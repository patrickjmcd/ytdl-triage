// ytdl-triage serves a small web UI for manually resolving low-confidence
// artist/title guesses the youtubedl postprocessor couldn't make on its own
// — see https://github.com/patrickjmcd/k8s-argo for the pipeline this feeds.
package main

import (
	"log"
	"net/http"

	"github.com/patrickjmcd/ytdl-triage/internal/api"
	"github.com/patrickjmcd/ytdl-triage/internal/config"
	"github.com/patrickjmcd/ytdl-triage/internal/web"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	api.NewServer(cfg).Routes(mux)

	frontend, err := web.Handler()
	if err != nil {
		log.Fatalf("loading embedded frontend: %v", err)
	}
	mux.Handle("/", frontend)

	log.Printf("ytdl-triage listening on :%s (needsReviewDir=%s organizedDir=%s)", cfg.Port, cfg.NeedsReviewDir, cfg.OrganizedDir)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
